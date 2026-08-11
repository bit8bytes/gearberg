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
package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	gencs "github.com/bit8bytes/gearberg/internal/database/queries/gen/orgsettings"
	"github.com/bit8bytes/gearberg/internal/money"
	"github.com/bit8bytes/gearberg/internal/timezone"
)

// Repository provides data access for org settings.
type Repository struct {
	db       *sql.DB
	settings gencs.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:       db,
		settings: gencs.New(db),
	}
}

// GetByOrgID returns the settings for orgID, or nil when none exist yet.
func (r *Repository) GetByOrgID(ctx context.Context, orgID string) (*OrgSettings, error) {
	row, err := r.settings.GetByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByOrgID: %w", err)
	}
	return toModel(row), nil
}

// Create inserts a new settings row with migration defaults for currency, vat_rate, and timezone.
func (r *Repository) Create(ctx context.Context, id, orgID string) (*OrgSettings, error) {
	row, err := r.settings.Create(ctx, gencs.CreateParams{
		ID:    id,
		OrgID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return &OrgSettings{
		ID:       row.ID,
		OrgID:    row.OrgID,
		Currency: money.Currency(row.Currency),
		VatRate:  money.VatRate(row.VatRate),
		Timezone: timezone.Timezone(row.Timezone),
	}, nil
}

// Update applies currency, vat_rate, and timezone changes to the existing settings row for orgID.
func (r *Repository) Update(ctx context.Context, u UpdateOrgSettings) (*OrgSettings, error) {
	existing, err := r.settings.GetByOrgID(ctx, u.OrgID)
	if err != nil {
		return nil, fmt.Errorf("Update: %w", err)
	}
	row, err := r.settings.Update(ctx, gencs.UpdateParams{
		ID:       existing.ID,
		Currency: u.Currency.String(),
		VatRate:  int64(u.VatRate),
		Timezone: u.Timezone.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("Update: %w", err)
	}
	return toModel(row), nil
}

func toModel(s gencs.OrgSetting) *OrgSettings {
	return &OrgSettings{
		ID:       s.ID,
		OrgID:    s.OrgID,
		Currency: money.Currency(s.Currency),
		VatRate:  money.VatRate(s.VatRate),
		Timezone: timezone.Timezone(s.Timezone),
	}
}
