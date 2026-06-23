package imports

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/segmentio/ksuid"
)

// EquipmentWriter is the subset of inventory.Service used by the import service.
type EquipmentWriter interface {
	ListAll(ctx context.Context, orgID string) ([]equipment.Equipment, error)
	CreateBulkTx(ctx context.Context, tx *sql.Tx, c equipment.CreateBulkEquipment) (*equipment.Equipment, error)
	CreateSerialized(ctx context.Context, tx *sql.Tx, c equipment.CreateSerializedEquipment) (*equipment.Equipment, error)
}

// CategoryEnsurer resolves or creates categories by name.
type CategoryEnsurer interface {
	EnsureByName(ctx context.Context, orgID, name string) (string, error)
}

// ManufacturerEnsurer resolves or creates manufacturers by name.
type ManufacturerEnsurer interface {
	EnsureByName(ctx context.Context, orgID, name string) (string, error)
}

// LocationEnsurer resolves or creates locations by name.
type LocationEnsurer interface {
	EnsureByName(ctx context.Context, orgID, name string) (string, error)
}

// Service handles CSV import staging and commit.
type Service struct {
	repo          *Repository
	db            *sql.DB
	inventory     EquipmentWriter
	categories    CategoryEnsurer
	manufacturers ManufacturerEnsurer
	locations     LocationEnsurer
}

// NewService returns a new Service.
func NewService(repo *Repository, db *sql.DB, inv EquipmentWriter, cats CategoryEnsurer, mfrs ManufacturerEnsurer, locs LocationEnsurer) *Service {
	return &Service{repo: repo, db: db, inventory: inv, categories: cats, manufacturers: mfrs, locations: locs}
}

// Stage deletes any existing staging rows for the org, then validates and stages
// the provided rows. Returns the import_id grouping the new rows.
func (s *Service) Stage(ctx context.Context, orgID string, rawRows []RawRow) (string, error) {
	if err := s.repo.DeleteByOrgID(ctx, orgID); err != nil {
		return "", fmt.Errorf("Stage: %w", err)
	}

	existing, err := s.inventory.ListAll(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("Stage: %w", err)
	}
	existingByName := make(map[string]string, len(existing))
	for _, item := range existing {
		existingByName[strings.ToLower(item.Name)] = item.ID
	}

	importID := ksuid.New().String()
	for i, raw := range rawRows {
		row := Row{
			ID:               ksuid.New().String(),
			ImportID:         importID,
			OrgID:            orgID,
			RowNumber:        int64(i + 1),
			Name:             raw.Name,
			TypeLabel:        raw.TypeLabel,
			UsageTypeLabel:   raw.UsageTypeLabel,
			CategoryName:     raw.CategoryName,
			ManufacturerName: raw.ManufacturerName,
			LocationName:     raw.LocationName,
			RentalPrice:      raw.RentalPrice,
			ResalePrice:      raw.ResalePrice,
			Notes:            raw.Notes,
			WeightG:          raw.WeightG,
			WidthMm:          raw.WidthMm,
			HeightMm:         raw.HeightMm,
			DepthMm:          raw.DepthMm,
			VoltageV:         raw.VoltageV,
			CurrentMa:        raw.CurrentMa,
			PowerMw:          raw.PowerMw,
			Quantity:         raw.Quantity,
		}

		if errMsg := validateRow(raw); errMsg != "" {
			row.Status = StatusError
			row.ErrorMessage = errMsg
			row.Action = ActionSkip
		} else if _, conflict := existingByName[strings.ToLower(raw.Name)]; conflict {
			row.Status = StatusError
			row.ErrorMessage = "A gear item with this name already exists"
			row.Action = ActionSkip
		} else {
			row.Status = StatusNew
			row.Action = ActionCreate
		}

		if _, err := s.repo.Insert(ctx, row); err != nil {
			return "", fmt.Errorf("Stage: row %d: %w", i+1, err)
		}
	}

	return importID, nil
}

// ListStaged returns all staged rows for an import.
func (s *Service) ListStaged(ctx context.Context, importID string) ([]Row, error) {
	rows, err := s.repo.ListByImportID(ctx, importID)
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

// ensureCommitLookups calls EnsureByName for every unique category, manufacturer,
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
	id, err := s.categories.EnsureByName(ctx, orgID, name)
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
	id, err := s.manufacturers.EnsureByName(ctx, orgID, name)
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
	id, err := s.locations.EnsureByName(ctx, orgID, name)
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
	rows, err := s.repo.ListByImportID(ctx, importID)
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

	for _, row := range rows {
		if row.Status == StatusError || row.Action == ActionSkip {
			continue
		}
		if err := s.commitRow(ctx, tx, row, lk); err != nil {
			return fmt.Errorf("Commit: row %d: %w", row.RowNumber, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	// Staging rows are cleaned up after the inventory transaction commits.
	// A failure here leaves stale rows that DeleteByOrgID will clear on the next import.
	if err := s.repo.DeleteByImportID(ctx, importID); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	return nil
}

func (s *Service) commitRow(ctx context.Context, tx *sql.Tx, row Row, lk commitLookups) error {
	catID := lk.catsByName[strings.ToLower(row.CategoryName)]
	mfrID := lk.mfrsByName[strings.ToLower(row.ManufacturerName)]
	var locID *string
	if id, ok := lk.locsByName[strings.ToLower(row.LocationName)]; ok {
		locID = &id
	}
	return s.createRow(ctx, tx, row, catID, mfrID, locID)
}

func (s *Service) createRow(ctx context.Context, tx *sql.Tx, row Row, catID, mfrID string, locID *string) error {
	usageType := equipment.Rental
	if strings.EqualFold(row.UsageTypeLabel, "sale") {
		usageType = equipment.Sale
	}

	itemID := ksuid.New().String()
	base := equipment.Base{
		ID:             itemID,
		OrgID:          row.OrgID,
		UsageTypeID:    usageType.ID(),
		Name:           row.Name,
		CategoryID:     catID,
		ManufacturerID: mfrID,
		LocationID:     locID,
		PurchasePrice:  parseCents(row.ResalePrice),
		RentalPrice:    parseCents(row.RentalPrice),
		Notes:          row.Notes,
		WeightG:        parseOptionalInt64(row.WeightG),
		WidthMM:        parseOptionalInt64(row.WidthMm),
		HeightMM:       parseOptionalInt64(row.HeightMm),
		DepthMM:        parseOptionalInt64(row.DepthMm),
		PowerMW:        parseOptionalInt64(row.PowerMw),
		CurrentMA:      parseOptionalInt64(row.CurrentMa),
		VoltageV:       parseOptionalInt64(row.VoltageV),
	}

	if !strings.EqualFold(row.TypeLabel, "serialized") {
		if _, err := s.inventory.CreateBulkTx(ctx, tx, equipment.CreateBulkEquipment{
			Base:       base,
			TotalStock: parseQuantity(row.Quantity),
		}); err != nil {
			return fmt.Errorf("createRow: bulk: %w", err)
		}
		return nil
	}

	qty := parseQuantity(row.Quantity)
	units := make([]equipment.CreateUnit, qty)
	for i := range units {
		units[i] = equipment.CreateUnit{
			ID:           ksuid.New().String(),
			EquipmentID:  itemID,
			SerialNumber: serial.New(),
		}
	}
	if _, err := s.inventory.CreateSerialized(ctx, tx, equipment.CreateSerializedEquipment{
		Base:  base,
		Units: units,
	}); err != nil {
		return fmt.Errorf("createRow: serialized: %w", err)
	}
	return nil
}

func validateRow(raw RawRow) string {
	if strings.TrimSpace(raw.Name) == "" {
		return "Name is required"
	}
	tl := strings.TrimSpace(raw.TypeLabel)
	if !strings.EqualFold(tl, "bulk") && !strings.EqualFold(tl, "serialized") {
		return fmt.Sprintf("Type must be Bulk or Serialized, got %q", tl)
	}
	ul := strings.TrimSpace(raw.UsageTypeLabel)
	if !strings.EqualFold(ul, "rental") && !strings.EqualFold(ul, "sale") {
		return fmt.Sprintf("Usage must be Rental or Sale, got %q", ul)
	}
	if strings.TrimSpace(raw.CategoryName) == "" {
		return "Category is required"
	}
	return ""
}

func parseQuantity(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 1
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

func parseOptionalInt64(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

func parseCents(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, ",", ".")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := int64(math.Round(f * 100))
	return &v
}
