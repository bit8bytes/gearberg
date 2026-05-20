package categories

import (
	"context"
	"errors"
	"fmt"
	"html/template"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
	"github.com/tobiasgleiter/forma"
)

// Handler holds dependencies for equipment category HTTP handlers.
type Handler struct {
	svc   *Service
	cache map[string]*template.Template
}

// NewHandler returns a new Handler wired with svc.
func NewHandler(svc *Service, cache map[string]*template.Template) *Handler {
	return &Handler{svc: svc, cache: cache}
}

// Routes registers equipment category HTML form routes on m.
func (h *Handler) Routes(m *forma.HTML) {
	forma.Get(m, forma.Operation{
		Path:     "/companies/{company_id}/settings/equipment-categories",
		Template: h.cache[pages.EquipmentCategoriesIndex.File],
	}, h.listCategories)

	forma.Get(m, forma.Operation{
		Path:     "/companies/{company_id}/settings/equipment-categories/new",
		Template: h.cache[pages.EquipmentCategoriesNew.File],
	}, h.getCategoryNew)

	forma.Postf(m,
		forma.Operation{
			Path:     "/companies/{company_id}/settings/equipment-categories/new",
			Template: h.cache[pages.EquipmentCategoriesNew.File],
		},
		func(o *categoryFormOutput) string {
			return fmt.Sprintf("/companies/%s/settings/equipment-categories", o.CompanyID)
		}, h.createCategory)

	forma.Get(m, forma.Operation{
		Path:     "/companies/{company_id}/settings/equipment-categories/{id}",
		Template: h.cache[pages.EquipmentCategoriesDetail.File],
	}, h.getCategory)

	forma.Post(m, forma.Operation{
		Path:     "/companies/{company_id}/settings/equipment-categories/{id}",
		Template: h.cache[pages.EquipmentCategoriesDetail.File],
	}, h.updateCategory)

	forma.Postf(m, forma.Operation{
		Path:     "/companies/{company_id}/settings/equipment-categories/{id}/delete",
		Template: h.cache[pages.EquipmentCategoriesDetail.File],
	},
		func(o *categoryFormOutput) string {
			return fmt.Sprintf("/companies/%s/settings/equipment-categories", o.CompanyID)
		}, h.deleteCategory)
}

type listCategoriesInput struct {
	CompanyID string `path:"company_id"`
}

type listCategoriesOutput struct {
	CompanyID     string
	Categories    []EquipmentCategory
	MaxCategories int
}

func (h *Handler) listCategories(ctx context.Context, in *listCategoriesInput) (*listCategoriesOutput, error) {
	equipmentCategories, err := h.svc.GetByCompanyID(ctx, in.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("listCategories: %w", err)
	}
	return &listCategoriesOutput{
		CompanyID:     in.CompanyID,
		Categories:    equipmentCategories,
		MaxCategories: h.svc.MaxCategories(),
	}, nil
}

type categoryFormInput struct {
	CompanyID string `path:"company_id"`
	Name      string `form:"name" required:"true" maxLength:"100"`
}

type categoryFormOutput struct {
	CompanyID string
	Category  *EquipmentCategory
	redirect  string
}

func (o *categoryFormOutput) RedirectURL() string { return o.redirect }

func (h *Handler) getCategoryNew(_ context.Context, in *categoryFormInput) (*categoryFormOutput, error) {
	return &categoryFormOutput{CompanyID: in.CompanyID}, nil
}

func (h *Handler) createCategory(ctx context.Context, in *categoryFormInput) (*categoryFormOutput, error) {
	_, err := h.svc.Create(ctx, CreateEquipmentCategory{
		ID:        ksuid.New().String(),
		CompanyID: in.CompanyID,
		Name:      in.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			return nil, &forma.ValidationError{Field: map[string]string{"name": "A category with this name already exists."}}
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := h.svc.MaxCategories()
			return nil, &forma.ValidationError{Field: map[string]string{"name": fmt.Sprintf("Category limit reached. Only %d categories allowed per company.", limit)}}
		}
		return nil, fmt.Errorf("createCategory: %w", err)
	}
	return &categoryFormOutput{CompanyID: in.CompanyID}, nil
}

type categoryDetailInput struct {
	CompanyID string `path:"company_id"`
	ID        string `path:"id"`
	Name      string `form:"name" required:"true" maxLength:"100"`
}

func (h *Handler) getCategory(ctx context.Context, in *categoryDetailInput) (*categoryFormOutput, error) {
	category, err := h.svc.GetByID(ctx, in.ID)
	if err != nil {
		return nil, fmt.Errorf("getCategory: %w", err)
	}
	return &categoryFormOutput{CompanyID: in.CompanyID, Category: category}, nil
}

func (h *Handler) updateCategory(ctx context.Context, in *categoryDetailInput) (*categoryFormOutput, error) {
	category, err := h.svc.Update(ctx, UpdateEquipmentCategory{
		ID:   in.ID,
		Name: in.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("updateCategory: %w", err)
	}
	return &categoryFormOutput{CompanyID: in.CompanyID, Category: category}, nil
}

type deleteCategoryInput struct {
	CompanyID string `path:"company_id"`
	ID        string `path:"id"`
}

func (h *Handler) deleteCategory(ctx context.Context, in *deleteCategoryInput) (*categoryFormOutput, error) {
	if err := h.svc.Delete(ctx, in.ID); err != nil {
		return nil, fmt.Errorf("deleteCategory: %w", err)
	}
	return &categoryFormOutput{CompanyID: in.CompanyID}, nil
}
