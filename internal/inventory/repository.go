package inventory

import (
	"context"
	"database/sql"
	"fmt"

	geninv "github.com/bit8bytes/gearberg/database/queries/gen/inventory"
)

// Repository provides data access for inventory items.
type Repository struct {
	db        *sql.DB
	inventory geninv.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:        db,
		inventory: geninv.New(db),
	}
}

// GetByCompanyID returns all inventory items belonging to companyID.
func (r *Repository) GetByCompanyID(ctx context.Context, companyID string) ([]Inventory, error) {
	rows, err := r.inventory.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("GetByCompanyID: %w", err)
	}
	result := make([]Inventory, len(rows))
	for i, row := range rows {
		result[i] = Inventory{
			ID:         row.ID,
			Name:       row.Name,
			CategoryID: row.CategoryID,
			TotalStock: row.TotalStock,
		}
	}
	return result, nil
}
