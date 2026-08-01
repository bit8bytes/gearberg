// Package settings handles org settings routes, business logic, and data access.
package settings

import (
	"github.com/bit8bytes/gearberg/internal/money"
	"github.com/bit8bytes/gearberg/internal/timezone"
)

// OrgSettings holds the configurable settings for a org.
type OrgSettings struct {
	ID       string
	OrgID    string
	Currency money.Currency
	VatRate  money.VatRate
	Timezone timezone.Timezone
}

// UpdateOrgSettings holds the data required to update org settings.
type UpdateOrgSettings struct {
	OrgID    string
	Currency money.Currency
	VatRate  money.VatRate
	Timezone timezone.Timezone
}
