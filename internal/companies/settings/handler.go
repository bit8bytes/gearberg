package settings

import (
	"context"
	"fmt"
	"html/template"

	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
	"github.com/tobiasgleiter/forma"
)

// Handler holds dependencies for company settings HTTP handlers.
type Handler struct {
	svc   *Service
	cache map[string]*template.Template
}

// NewHandler returns a new Handler wired with svc.
func NewHandler(svc *Service, cache map[string]*template.Template) *Handler {
	return &Handler{svc: svc, cache: cache}
}

// Routes registers company settings routes on m.
func (h *Handler) Routes(m *forma.HTML) {
	forma.Get(m, forma.Operation{
		Path:     "/companies/{company_id}/settings",
		Template: h.cache[pages.CompanySettingsConfig.File],
	}, h.getSettings)

	forma.Post(m, forma.Operation{
		Path:     "/companies/{company_id}/settings",
		Template: h.cache[pages.CompanySettingsConfig.File],
	}, h.upsertSettings)
}

type settingsInput struct {
	CompanyID string  `path:"company_id"`
	Currency  string  `form:"currency" required:"true" iso:"4217"`
	VatRate   float64 `form:"vat_rate" min:"0.00" max:"1.00"`
	Timezone  string  `form:"timezone" required:"true" timezone:"iana"`
}

type settingsOutput struct {
	Settings *CompanySettings
}

func (h *Handler) getSettings(ctx context.Context, in *settingsInput) (*settingsOutput, error) {
	settings, err := h.svc.GetByCompanyID(ctx, in.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("getSettings: %w", err)
	}
	return &settingsOutput{Settings: settings}, nil
}

func (h *Handler) upsertSettings(ctx context.Context, in *settingsInput) (*settingsOutput, error) {
	settings, err := h.svc.Upsert(ctx, UpsertCompanySettings{
		ID:        ksuid.New().String(),
		CompanyID: in.CompanyID,
		Currency:  in.Currency,
		VatRate:   in.VatRate,
		Timezone:  in.Timezone,
	})
	if err != nil {
		return nil, fmt.Errorf("upsertSettings: %w", err)
	}
	return &settingsOutput{Settings: settings}, nil
}
