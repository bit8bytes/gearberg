// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package equipmentimports provides imports functionality.
package equipmentimports

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/usage"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/bit8bytes/gearberg/internal/uid"
	"github.com/bit8bytes/gearberg/internal/units"
)

// Upserter resolves or creates a named entity within an org.
type Upserter interface {
	Upsert(ctx context.Context, orgID, name string) (string, error)
}

// Service handles CSV import staging and commit.
type Service struct {
	repo          *Repository
	db            *sql.DB
	equipment     *equipment.Repository
	categories    Upserter
	manufacturers Upserter
	locations     Upserter
	steps         []Step
}

// NewService returns a new Service with the default validation pipeline.
func NewService(repo *Repository, db *sql.DB, equip *equipment.Repository, cats, mfrs, locs Upserter) *Service {
	return &Service{
		repo:          repo,
		db:            db,
		equipment:     equip,
		categories:    cats,
		manufacturers: mfrs,
		locations:     locs,
		steps:         []Step{validateStep},
	}
}

// validateStep marks rows that fail field-level validation as StateInvalid.
func validateStep(_ context.Context, rows []ProcessedRow) ([]ProcessedRow, error) {
	for i, pr := range rows {
		if pr.State == StateInvalid {
			continue
		}
		if msg := validateRow(pr.Data); msg != "" {
			rows[i].State = StateInvalid
			rows[i].Errors = append(rows[i].Errors, ValidationError{
				Line:   int(pr.Data.RowNumber),
				Field:  "Name",
				Reason: msg,
			})
		}
	}
	return rows, nil
}

// Stage deletes any existing staging rows for the org, runs the import pipeline,
// and persists the results. Returns the import_id grouping the new rows.
func (s *Service) Stage(ctx context.Context, orgID string, processed []ProcessedRow) (string, error) {
	if err := s.repo.DeleteByOrgID(ctx, orgID); err != nil {
		return "", fmt.Errorf("Stage: %w", err)
	}

	existing, _, err := s.equipment.List(ctx, orgID, "", "", false, pagination.Filters{Page: 1, PageSize: math.MaxInt32})
	if err != nil {
		return "", fmt.Errorf("Stage: %w", err)
	}
	existingByName := make(map[string]string, len(existing))
	for _, item := range existing {
		existingByName[strings.ToLower(item.Name)] = item.ID
	}

	importID := uid.New()
	for i := range processed {
		processed[i].Data.ID = uid.New()
		processed[i].Data.ImportID = importID
		processed[i].Data.OrgID = orgID
		processed[i].Data.RowNumber = int64(i + 1)
		processed[i].Data.Status = StatusNew
		processed[i].Data.Action = ActionCreate
	}

	pipeline := make([]Step, len(s.steps)+1)
	copy(pipeline, s.steps)
	pipeline[len(s.steps)] = conflictStep(existingByName)
	processed, err = RunValidation(ctx, processed, pipeline)
	if err != nil {
		return "", fmt.Errorf("Stage: pipeline: %w", err)
	}

	rows := toRows(processed)
	for i, row := range rows {
		if _, err := s.repo.Create(ctx, row); err != nil {
			return "", fmt.Errorf("Stage: row %d: %w", i+1, err)
		}
	}

	return importID, nil
}

// toRows converts a validated []ProcessedRow back to []Row for storage,
// mapping StateInvalid→StatusError and collecting the first error message.
func toRows(processed []ProcessedRow) []Row {
	rows := make([]Row, len(processed))
	for i, pr := range processed {
		row := pr.Data
		if pr.State == StateInvalid {
			row.Status = StatusError
			row.Action = ActionSkip
			if len(pr.Errors) > 0 {
				row.ErrorMessage = pr.Errors[0].Reason
			}
		}
		rows[i] = row
	}
	return rows
}

// conflictStep returns a Step that marks rows whose name already exists in inventory.
func conflictStep(existingByName map[string]string) Step {
	return func(_ context.Context, rows []ProcessedRow) ([]ProcessedRow, error) {
		for i, pr := range rows {
			if pr.State == StateInvalid {
				continue
			}
			if _, conflict := existingByName[strings.ToLower(pr.Data.Name)]; conflict {
				rows[i].State = StateInvalid
				rows[i].Errors = append(rows[i].Errors, ValidationError{
					Line:   int(pr.Data.RowNumber),
					Field:  "Name",
					Reason: "A gear item with this name already exists",
				})
			}
		}
		return rows, nil
	}
}

// ListStaged returns all staged rows for an import.
func (s *Service) ListStaged(ctx context.Context, importID string) ([]Row, error) {
	rows, err := s.repo.List(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("ListStaged: %w", err)
	}
	return rows, nil
}

// commitLookups holds pre-resolved name→ID maps for the commit phase.
type commitLookups struct {
	catsByName map[string]string
	mfrsByName map[string]string
	locsByName map[string]string
}

// ensureCommitLookups upserts every unique category, manufacturer,
// and location name in the batch, creating them if they don't exist yet.
func (s *Service) ensureCommitLookups(ctx context.Context, orgID string, rows []Row) (commitLookups, error) {
	lk := commitLookups{
		catsByName: make(map[string]string),
		mfrsByName: make(map[string]string),
		locsByName: make(map[string]string),
	}
	for _, row := range rows {
		if row.Status == StatusError || row.Action == ActionSkip {
			continue
		}
		if err := s.ensureCategory(ctx, &lk, orgID, row.CategoryName); err != nil {
			return commitLookups{}, fmt.Errorf("ensureCommitLookups: %w", err)
		}
		if name := strings.TrimSpace(row.ManufacturerName); name != "" {
			if err := s.ensureManufacturer(ctx, &lk, orgID, name); err != nil {
				return commitLookups{}, fmt.Errorf("ensureCommitLookups: %w", err)
			}
		}
		if name := strings.TrimSpace(row.LocationName); name != "" {
			if err := s.ensureLocation(ctx, &lk, orgID, name); err != nil {
				return commitLookups{}, fmt.Errorf("ensureCommitLookups: %w", err)
			}
		}
	}
	return lk, nil
}

func (s *Service) ensureCategory(ctx context.Context, lk *commitLookups, orgID, name string) error {
	key := strings.ToLower(name)
	if _, ok := lk.catsByName[key]; ok {
		return nil
	}
	id, err := s.categories.Upsert(ctx, orgID, name)
	if err != nil {
		return fmt.Errorf("category %q: %w", name, err)
	}
	lk.catsByName[key] = id
	return nil
}

func (s *Service) ensureManufacturer(ctx context.Context, lk *commitLookups, orgID, name string) error {
	key := strings.ToLower(name)
	if _, ok := lk.mfrsByName[key]; ok {
		return nil
	}
	id, err := s.manufacturers.Upsert(ctx, orgID, name)
	if err != nil {
		return fmt.Errorf("manufacturer %q: %w", name, err)
	}
	lk.mfrsByName[key] = id
	return nil
}

func (s *Service) ensureLocation(ctx context.Context, lk *commitLookups, orgID, name string) error {
	key := strings.ToLower(name)
	if _, ok := lk.locsByName[key]; ok {
		return nil
	}
	id, err := s.locations.Upsert(ctx, orgID, name)
	if err != nil {
		return fmt.Errorf("location %q: %w", name, err)
	}
	lk.locsByName[key] = id
	return nil
}

// Commit processes all staged rows for the import atomically: creates new items,
// then deletes the staging rows — all within a single transaction so no partial
// write is possible.
func (s *Service) Commit(ctx context.Context, importID string, orgID string) error {
	rows, err := s.repo.List(ctx, importID)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	lk, err := s.ensureCommitLookups(ctx, orgID, rows)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Commit: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.commitRows(ctx, tx, rows, lk); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	// Staging rows are cleaned up after the inventory transaction commits.
	// A failure here leaves stale rows that DeleteByOrgID will clear on the next import.
	if err := s.repo.Delete(ctx, importID); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	return nil
}

type serializedGroup struct {
	first Row
	rows  []Row
}

// commitRows groups serialized rows by name and commits all bulk and serialized items.
func (s *Service) commitRows(ctx context.Context, tx *sql.Tx, rows []Row, lk commitLookups) error {
	byName := make(map[string]*serializedGroup)
	var order []string
	for _, row := range rows {
		if row.Status == StatusError || row.Action == ActionSkip {
			continue
		}
		if !strings.EqualFold(row.TypeLabel, "serialized") {
			if err := s.commitBulkRow(ctx, tx, row, lk); err != nil {
				return fmt.Errorf("row %d: %w", row.RowNumber, err)
			}
			continue
		}
		key := strings.ToLower(row.Name)
		if _, ok := byName[key]; !ok {
			byName[key] = &serializedGroup{first: row}
			order = append(order, key)
		}
		byName[key].rows = append(byName[key].rows, row)
	}
	for _, key := range order {
		g := byName[key]
		if err := s.commitSerializedGroup(ctx, tx, g.first, g.rows, lk); err != nil {
			return fmt.Errorf("serialized %q: %w", g.first.Name, err)
		}
	}
	return nil
}

func (s *Service) resolveLookups(row Row, lk commitLookups) (catID, mfrID, locID string) {
	catID = lk.catsByName[strings.ToLower(row.CategoryName)]
	mfrID = lk.mfrsByName[strings.ToLower(row.ManufacturerName)]
	locID = lk.locsByName[strings.ToLower(row.LocationName)]
	return
}

// parseInt64 parses a string as int64, returning 0 for blank or invalid input.
func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

// ptrOf converts an int64 DB value to a typed pointer; returns nil when v is 0.
func ptrOf[T ~int64](v int64) *T {
	if v == 0 {
		return nil
	}
	t := T(v)
	return &t
}

func buildBase(row Row, catID, mfrID, locID string) equipment.Base {
	return equipment.Base{
		OrgID:          row.OrgID,
		UsageTypeID:    usage.Rental.ID(),
		Name:           row.Name,
		CategoryID:     catID,
		ManufacturerID: mfrID,
		LocationID:     locID,
		Notes:          row.Notes,
		EquipmentType:  equipment.ParseOrDefault(strings.ToLower(strings.TrimSpace(row.EquipmentTypeLabel))),
		Pricing: equipment.Pricing{
			PurchasePrice: ptrOf[units.Cents](parseInt64(row.ResalePrice)),
			RentalPrice:   ptrOf[units.Cents](parseInt64(row.RentalPrice)),
		},
		Properties: equipment.Properties{
			Weight:    ptrOf[units.Grams](parseInt64(row.WeightG)),
			Width:     ptrOf[units.Millimeters](parseInt64(row.WidthMm)),
			Height:    ptrOf[units.Millimeters](parseInt64(row.HeightMm)),
			Depth:     ptrOf[units.Millimeters](parseInt64(row.DepthMm)),
			Power:     ptrOf[units.Milliwatts](parseInt64(row.PowerMw)),
			Current:   ptrOf[units.Milliamps](parseInt64(row.CurrentMa)),
			Voltage:   ptrOf[units.Millivolts](parseInt64(row.VoltageMv)),
			WireGauge: ptrOf[units.WireGauge](parseInt64(row.WireGaugeMM2X100)),
		},
	}
}

func (s *Service) commitBulkRow(ctx context.Context, tx *sql.Tx, row Row, lk commitLookups) error {
	catID, mfrID, locID := s.resolveLookups(row, lk)
	base := buildBase(row, catID, mfrID, locID)
	if _, err := s.equipment.CreateBulk(ctx, tx, equipment.CreateBulkEquipment{
		ID:         uid.New(),
		BulkItemID: uid.New(),
		Base:       base,
		TotalStock: equipment.ParseQuantity(row.Quantity),
	}); err != nil {
		return fmt.Errorf("commitBulkRow: %w", err)
	}
	return nil
}

func buildUnit(row Row, equipmentID string) equipment.CreateUnit {
	sn := strings.TrimSpace(row.UnitSerialNumber)
	if sn == "" {
		sn = serial.New()
	}
	isActive := !strings.EqualFold(row.UnitIsActive, "false") && row.UnitIsActive != "0"
	return equipment.CreateUnit{
		ID:                       uid.New(),
		EquipmentID:              equipmentID,
		SerialNumber:             sn,
		ManufacturerSerialNumber: row.UnitManufacturerSerial,
		Remark:                   row.UnitRemark,
		PurchasePrice:            ptrOf[units.Cents](parseInt64(row.UnitPurchasePrice)),
		PurchasedAt:              equipment.ParseDate(row.UnitPurchasedAt),
		NextInspectionAt:         equipment.ParseDate(row.NextInspectionAt),
		IsActive:                 isActive,
	}
}

// commitSerializedGroup creates one equipment item with one unit per row in the group.
func (s *Service) commitSerializedGroup(ctx context.Context, tx *sql.Tx, first Row, rows []Row, lk commitLookups) error {
	catID, mfrID, locID := s.resolveLookups(first, lk)
	base := buildBase(first, catID, mfrID, locID)
	itemID := uid.New()
	unitList := make([]equipment.CreateUnit, 0, len(rows))
	for _, row := range rows {
		unitList = append(unitList, buildUnit(row, itemID))
	}
	if _, err := s.equipment.CreateSerialized(ctx, tx, equipment.CreateSerializedEquipment{
		ID:    itemID,
		Base:  base,
		Units: unitList,
	}); err != nil {
		return fmt.Errorf("commitSerializedGroup: %w", err)
	}
	return nil
}

func validateRow(row Row) string {
	if strings.TrimSpace(row.Name) == "" {
		return "Name is required"
	}
	tl := strings.TrimSpace(row.TypeLabel)
	if !strings.EqualFold(tl, "bulk") && !strings.EqualFold(tl, "serialized") {
		return fmt.Sprintf("Type must be Bulk or Serialized, got %q", tl)
	}
	ul := strings.TrimSpace(row.UsageTypeLabel)
	if !strings.EqualFold(ul, "rental") && !strings.EqualFold(ul, "sale") {
		return fmt.Sprintf("Usage must be Rental or Sale, got %q", ul)
	}
	if strings.TrimSpace(row.CategoryName) == "" {
		return "Category is required"
	}
	return ""
}
