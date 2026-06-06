package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	gencs "github.com/bit8bytes/gearberg/internal/database/queries/gen/orgsettings"
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

// Upsert creates settings when none exist for the org, or updates the existing row.
// The ID field in u is used only on insert; on update the existing row's ID is kept.
func (r *Repository) Upsert(ctx context.Context, u UpsertOrgSettings) (*OrgSettings, error) {
	existing, err := r.settings.GetByOrgID(ctx, u.OrgID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("Upsert: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		row, err := r.settings.Create(ctx, gencs.CreateParams{
			ID:       u.ID,
			OrgID:    u.OrgID,
			Currency: u.Currency,
			VatRate:  u.VatRate,
			Timezone: u.Timezone,
		})
		if err != nil {
			return nil, fmt.Errorf("Upsert: %w", err)
		}
		return &OrgSettings{
			ID:       row.ID,
			OrgID:    row.OrgID,
			Currency: row.Currency,
			VatRate:  row.VatRate,
			Timezone: row.Timezone,
		}, nil
	}

	row, err := r.settings.Update(ctx, gencs.UpdateParams{
		ID:       existing.ID,
		Currency: u.Currency,
		VatRate:  u.VatRate,
		Timezone: u.Timezone,
	})
	if err != nil {
		return nil, fmt.Errorf("Upsert: %w", err)
	}
	return toModel(row), nil
}

func toModel(s gencs.OrgSetting) *OrgSettings {
	return &OrgSettings{
		ID:       s.ID,
		OrgID:    s.OrgID,
		Currency: s.Currency,
		VatRate:  s.VatRate,
		Timezone: s.Timezone,
	}
}
