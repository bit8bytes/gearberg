package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	gencs "github.com/bit8bytes/gearberg/database/queries/gen/companysettings"
)

// Repository provides data access for company settings.
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

// GetByCompanyID returns the settings for companyID, or nil when none exist yet.
func (r *Repository) GetByCompanyID(ctx context.Context, companyID string) (*CompanySettings, error) {
	row, err := r.settings.GetCompanySettingsByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByCompanyID: %w", err)
	}
	return toModel(row), nil
}

// Upsert creates settings when none exist for the company, or updates the existing row.
// The ID field in u is used only on insert; on update the existing row's ID is kept.
func (r *Repository) Upsert(ctx context.Context, u UpsertCompanySettings) (*CompanySettings, error) {
	existing, err := r.settings.GetCompanySettingsByCompanyID(ctx, u.CompanyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("Upsert: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		row, err := r.settings.CreateCompanySettings(ctx, gencs.CreateCompanySettingsParams{
			ID:        u.ID,
			CompanyID: u.CompanyID,
			Currency:  u.Currency,
			VatRate:   u.VatRate,
			Timezone:  u.Timezone,
		})
		if err != nil {
			return nil, fmt.Errorf("Upsert: %w", err)
		}
		return &CompanySettings{
			ID:        row.ID,
			CompanyID: row.CompanyID,
			Currency:  row.Currency,
			VatRate:   row.VatRate,
			Timezone:  row.Timezone,
		}, nil
	}

	row, err := r.settings.UpdateCompanySettings(ctx, gencs.UpdateCompanySettingsParams{
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

func toModel(s gencs.CompanySetting) *CompanySettings {
	return &CompanySettings{
		ID:        s.ID,
		CompanyID: s.CompanyID,
		Currency:  s.Currency,
		VatRate:   s.VatRate,
		Timezone:  s.Timezone,
	}
}
