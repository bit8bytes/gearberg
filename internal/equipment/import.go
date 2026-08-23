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

package equipment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bit8bytes/gearberg/internal/equipment/usage"
	imports "github.com/bit8bytes/gearberg/internal/import"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/bit8bytes/gearberg/internal/uid"
	"github.com/bit8bytes/gearberg/internal/units"
)

// Upserter resolves or creates a named entity within an org.
type Upserter interface {
	Upsert(ctx context.Context, orgID, name string) (string, error)
}

// ImportPipeline bundles the org-agnostic validation steps and commit handler
// for equipment imports. Build once at startup; orgID is supplied per request.
type ImportPipeline struct {
	equip         *Service
	categories    Upserter
	manufacturers Upserter
	locations     Upserter
}

// NewImportPipeline returns a stateless ImportPipeline. Wire it once in
// setupServices and inject it into the handler via the services struct.
func NewImportPipeline(equip *Service, cats, mfrs, locs Upserter) *ImportPipeline {
	return &ImportPipeline{
		equip:         equip,
		categories:    cats,
		manufacturers: mfrs,
		locations:     locs,
	}
}

// StepsFor returns the validation steps bound to orgID for a single session.
func (p *ImportPipeline) StepsFor(orgID string) []imports.Step {
	return []imports.Step{p.equip.conflictStep(orgID)}
}

// Commit implements imports.CommitHandler.
func (p *ImportPipeline) Commit(ctx context.Context, orgID string, rows []imports.Row, mappings []imports.Mapping) error {
	fieldFor := make(map[string]string, len(mappings))
	for _, m := range mappings {
		fieldFor[m.SourceCol] = m.TargetField
	}

	type serializedGroup struct {
		fields map[string]string
		units  []map[string]string
	}
	bulkRows := make([]map[string]string, 0)
	serializedByName := make(map[string]*serializedGroup)
	var serializedOrder []string

	for _, row := range rows {
		var raw map[string]string
		if err := json.Unmarshal([]byte(row.Data), &raw); err != nil {
			return fmt.Errorf("Commit: unmarshal row %d: %w", row.RowNumber, err)
		}
		fields := applyMappings(raw, fieldFor)

		typeLabel := strings.ToLower(strings.TrimSpace(fields["Type"]))
		if typeLabel == "serialized" {
			key := strings.ToLower(strings.TrimSpace(fields["Name"]))
			if _, ok := serializedByName[key]; !ok {
				serializedByName[key] = &serializedGroup{fields: fields}
				serializedOrder = append(serializedOrder, key)
			}
			serializedByName[key].units = append(serializedByName[key].units, fields)
		} else {
			bulkRows = append(bulkRows, fields)
		}
	}

	for _, fields := range bulkRows {
		if err := p.commitBulk(ctx, orgID, fields); err != nil {
			return fmt.Errorf("Commit: bulk %q: %w", fields["Name"], err)
		}
	}
	for _, key := range serializedOrder {
		g := serializedByName[key]
		if err := p.commitSerialized(ctx, orgID, g.fields, g.units); err != nil {
			return fmt.Errorf("Commit: serialized %q: %w", g.fields["Name"], err)
		}
	}
	return nil
}

func (p *ImportPipeline) commitBulk(ctx context.Context, orgID string, fields map[string]string) error {
	base, err := p.buildBase(ctx, orgID, fields)
	if err != nil {
		return err
	}
	qty := max(ParseQuantity(fields["Quantity"]), 1)
	_, err = p.equip.Create(ctx, CreateEquipment{
		Base:       base,
		ItemType:   "bulk",
		TotalStock: qty,
	})
	return err
}

func (p *ImportPipeline) commitSerialized(ctx context.Context, orgID string, fields map[string]string, unitFields []map[string]string) error {
	base, err := p.buildBase(ctx, orgID, fields)
	if err != nil {
		return err
	}
	itemID := uid.New()
	unitList := make([]CreateUnit, 0, len(unitFields))
	for _, uf := range unitFields {
		sn := strings.TrimSpace(uf["Unit Serial Number"])
		if sn == "" {
			sn = serial.New()
		}
		isActive := !strings.EqualFold(uf["Unit Active"], "false") && uf["Unit Active"] != "0"
		var purchasePrice *units.Cents
		if p := units.ParseCents(uf["Unit Purchase Price"]); p != nil {
			purchasePrice = p
		}
		unitList = append(unitList, CreateUnit{
			ID:                       uid.New(),
			OrgID:                    orgID,
			EquipmentID:              itemID,
			SerialNumber:             sn,
			ManufacturerSerialNumber: uf["Unit Manufacturer Serial"],
			Remark:                   uf["Unit Remark"],
			PurchasePrice:            purchasePrice,
			PurchasedAt:              ParseDate(uf["Unit Purchased At"]),
			NextInspectionAt:         ParseDate(uf["Next Inspection At"]),
			IsActive:                 isActive,
		})
	}
	_, err = p.equip.CreateWithUnits(ctx, base, unitList)
	return err
}

func (p *ImportPipeline) buildBase(ctx context.Context, orgID string, fields map[string]string) (Base, error) {
	catID, err := p.categories.Upsert(ctx, orgID, fields["Category"])
	if err != nil {
		return Base{}, fmt.Errorf("category %q: %w", fields["Category"], err)
	}
	var mfrID string
	if name := strings.TrimSpace(fields["Manufacturer"]); name != "" {
		mfrID, err = p.manufacturers.Upsert(ctx, orgID, name)
		if err != nil {
			return Base{}, fmt.Errorf("manufacturer %q: %w", name, err)
		}
	}
	var locID string
	if name := strings.TrimSpace(fields["Location"]); name != "" {
		locID, err = p.locations.Upsert(ctx, orgID, name)
		if err != nil {
			return Base{}, fmt.Errorf("location %q: %w", name, err)
		}
	}
	return Base{
		OrgID:          orgID,
		Name:           fields["Name"],
		UsageTypeID:    usage.Rental.ID(),
		CategoryID:     catID,
		ManufacturerID: mfrID,
		LocationID:     locID,
		Notes:          fields["Notes"],
		EquipmentType:  ParseOrDefault(strings.ToLower(strings.TrimSpace(fields["Equipment Type"]))),
		Pricing: Pricing{
			RentalPrice: units.ParseCents(fields["Rental Price"]),
		},
		Properties: Properties{
			Weight: units.ParseGrams(fields["Weight (kg)"]),
			Width:  units.ParseMillimeters(fields["Width (cm)"]),
			Height: units.ParseMillimeters(fields["Height (cm)"]),
			Depth:  units.ParseMillimeters(fields["Depth (cm)"]),
		},
	}, nil
}

// conflictStep returns a Step bound to orgID that marks rows whose "Name"
// already exists in inventory as an error.
// Serialized rows sharing a name within the same batch are not double-flagged.
func (s *Service) conflictStep(orgID string) imports.Step {
	return func(ctx context.Context, rows []imports.Row) ([]imports.Row, error) {
		existing, err := s.ListNames(ctx, orgID)
		if err != nil {
			return nil, fmt.Errorf("conflictStep: list names: %w", err)
		}
		nameSet := make(map[string]struct{}, len(existing))
		for _, n := range existing {
			nameSet[strings.ToLower(n)] = struct{}{}
		}

		for i, row := range rows {
			var fields map[string]string
			if err := json.Unmarshal([]byte(row.Data), &fields); err != nil {
				rows[i].Status = imports.RowStatusError
				rows[i].ErrorMessage = fmt.Sprintf("invalid JSON: %v", err)
				continue
			}

			name := strings.ToLower(strings.TrimSpace(fields["Name"]))
			if name == "" {
				rows[i].Status = imports.RowStatusError
				rows[i].ErrorMessage = "Name is required"
				continue
			}

			if _, conflict := nameSet[name]; conflict {
				rows[i].Status = imports.RowStatusError
				rows[i].ErrorMessage = fmt.Sprintf("equipment %q already exists in inventory", fields["Name"])
			}
		}
		return rows, nil
	}
}

func applyMappings(raw, fieldFor map[string]string) map[string]string {
	out := make(map[string]string, len(raw))
	for col, val := range raw {
		if target, ok := fieldFor[col]; ok {
			out[target] = val
		} else {
			out[col] = val
		}
	}
	return out
}
