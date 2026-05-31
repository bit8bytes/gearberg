// Package settings handles org settings routes, business logic, and data access.
package settings

// OrgSettings holds the configurable settings for a org.
type OrgSettings struct {
	ID       string
	OrgID    string
	Currency string
	VatRate  float64
	Timezone string
}

// UpsertOrgSettings holds the data required to create or update org settings.
type UpsertOrgSettings struct {
	ID       string
	OrgID    string
	Currency string
	VatRate  float64
	Timezone string
}
