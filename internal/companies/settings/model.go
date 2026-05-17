// Package settings handles company settings routes, business logic, and data access.
package settings

// CompanySettings holds the configurable settings for a company.
type CompanySettings struct {
	ID        string
	CompanyID string
	Currency  string
	VatRate   float64
	Timezone  string
}

// UpsertCompanySettings holds the data required to create or update company settings.
type UpsertCompanySettings struct {
	ID        string
	CompanyID string
	Currency  string
	VatRate   float64
	Timezone  string
}
