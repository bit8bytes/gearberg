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
		Currency: Currency(row.Currency),
		VatRate:  VatRate(row.VatRate),
		Timezone: Timezone(row.Timezone),
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
		Currency: Currency(s.Currency),
		VatRate:  VatRate(s.VatRate),
		Timezone: Timezone(s.Timezone),
	}
}
