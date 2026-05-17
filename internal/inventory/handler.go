// Package inventory handles inventory routes, business logic, and data access.
package inventory

import (
	"context"
	"fmt"
	"html/template"

	"github.com/bit8bytes/gearberg/internal/companies/categories"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/tobiasgleiter/forma"
)

// Handler holds dependencies for inventory HTTP handlers.
type Handler struct {
	svc   *Service
	cache map[string]*template.Template
}

// NewHandler returns a new Handler wired with svc.
func NewHandler(svc *Service, cache map[string]*template.Template) *Handler {
	return &Handler{svc: svc, cache: cache}
}

// Routes registers inventory routes on m.
func (h *Handler) Routes(m *forma.HTML) {
	forma.Get(m, forma.Operation{
		Path:     "/companies/{company_id}/inventory",
		Template: h.cache[pages.Inventory.File],
	}, h.listInventory)
}

type listInventoryInput struct {
	CompanyID string `path:"company_id"`
	Category  string `query:"category"`
}

type listInventoryOutput struct {
	Inventory  []Inventory
	Categories []categories.EquipmentCategory
}

func (h *Handler) listInventory(ctx context.Context, in *listInventoryInput) (*listInventoryOutput, error) {
	cats, err := h.svc.ListCategories(ctx, in.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("listInventory: %w", err)
	}
	return &listInventoryOutput{Categories: cats}, nil
}
