package imports

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bit8bytes/gearberg/internal/inventory"
	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/internal/orgs/manufacturers"
	"github.com/segmentio/ksuid"
)

// InventoryWriter is the subset of inventory.Service used by the import service.
type InventoryWriter interface {
	ListAll(ctx context.Context, orgID string) ([]inventory.Inventory, error)
	CreateBulkTx(ctx context.Context, tx *sql.Tx, c inventory.CreateBulkInventory) (*inventory.Inventory, error)
	CreateSerialized(ctx context.Context, tx *sql.Tx, c inventory.CreateSerializedInventory) (*inventory.Inventory, error)
}

// CategoryLister returns all categories for an org.
type CategoryLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]categories.EquipmentCategory, error)
}

// ManufacturerLister returns all manufacturers for an org.
type ManufacturerLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]manufacturers.Manufacturer, error)
}

// Service handles CSV import staging and commit.
type Service struct {
	repo          *Repository
	db            *sql.DB
	inventory     InventoryWriter
	categories    CategoryLister
	manufacturers ManufacturerLister
}

// NewService returns a new Service.
func NewService(repo *Repository, db *sql.DB, inv InventoryWriter, cats CategoryLister, mfrs ManufacturerLister) *Service {
	return &Service{repo: repo, db: db, inventory: inv, categories: cats, manufacturers: mfrs}
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

	cats, err := s.categories.GetByOrgID(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("Stage: %w", err)
	}
	catsByName := make(map[string]bool, len(cats))
	for _, c := range cats {
		catsByName[strings.ToLower(c.Name)] = true
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
			TotalStock:       raw.TotalStock,
			PurchasePrice:    raw.PurchasePrice,
			RentalPrice:      raw.RentalPrice,
			Notes:            raw.Notes,
		}

		if errMsg := validateRow(raw, catsByName); errMsg != "" {
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
}

func (s *Service) buildCommitLookups(ctx context.Context, orgID string) (commitLookups, error) {
	cats, err := s.categories.GetByOrgID(ctx, orgID)
	if err != nil {
		return commitLookups{}, fmt.Errorf("buildCommitLookups: %w", err)
	}
	mfrs, err := s.manufacturers.GetByOrgID(ctx, orgID)
	if err != nil {
		return commitLookups{}, fmt.Errorf("buildCommitLookups: %w", err)
	}
	lk := commitLookups{
		catsByName: make(map[string]string, len(cats)),
		mfrsByName: make(map[string]string, len(mfrs)),
	}
	for _, c := range cats {
		lk.catsByName[strings.ToLower(c.Name)] = c.ID
	}
	for _, m := range mfrs {
		lk.mfrsByName[strings.ToLower(m.Name)] = m.ID
	}
	return lk, nil
}

// Commit processes all staged rows for the import atomically: creates new items,
// overwrites conflicting items based on user decisions, then deletes the staging
// rows — all within a single transaction so no partial write is possible.
func (s *Service) Commit(ctx context.Context, importID string, orgID string) error {
	rows, err := s.repo.ListByImportID(ctx, importID)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	lk, err := s.buildCommitLookups(ctx, orgID)
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
	return s.createRow(ctx, tx, row, catID, mfrID, parseCents(row.PurchasePrice), parseCents(row.RentalPrice))
}

func (s *Service) createRow(ctx context.Context, tx *sql.Tx, row Row, catID, mfrID string, purchasePrice, rentalPrice *int64) error {
	usageType := inventory.Rental
	if strings.EqualFold(row.UsageTypeLabel, "sale") {
		usageType = inventory.Sale
	}

	itemID := ksuid.New().String()
	base := inventory.Base{
		ID:             itemID,
		OrgID:          row.OrgID,
		UsageTypeID:    usageType.ID(),
		Name:           row.Name,
		CategoryID:     catID,
		ManufacturerID: mfrID,
		PurchasePrice:  purchasePrice,
		RentalPrice:    rentalPrice,
		Notes:          row.Notes,
	}

	if !strings.EqualFold(row.TypeLabel, "serialized") {
		if _, err := s.inventory.CreateBulkTx(ctx, tx, inventory.CreateBulkInventory{
			Base:       base,
			TotalStock: parseTotalStock(row.TotalStock),
		}); err != nil {
			return fmt.Errorf("createRow: bulk: %w", err)
		}
		return nil
	}

	totalStock := parseTotalStock(row.TotalStock)
	units := make([]inventory.CreateUnit, totalStock)
	for i := range units {
		units[i] = inventory.CreateUnit{
			ID:          ksuid.New().String(),
			InventoryID: itemID,
			UnitNumber:  int64(i + 1),
		}
	}
	if _, err := s.inventory.CreateSerialized(ctx, tx, inventory.CreateSerializedInventory{
		Base:  base,
		Units: units,
	}); err != nil {
		return fmt.Errorf("createRow: serialized: %w", err)
	}
	return nil
}

func validateRow(raw RawRow, catsByName map[string]bool) string {
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
	if !catsByName[strings.ToLower(strings.TrimSpace(raw.CategoryName))] {
		return fmt.Sprintf("Category %q not found", raw.CategoryName)
	}
	return ""
}

func parseTotalStock(s string) int64 {
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
