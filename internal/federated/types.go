// Package federated manages federated identity records linking external OIDC
// providers to internal accounts.
package federated

// Provider represents an external identity provider. The integer value matches
// the id stored in the provider_types table, which is seeded at startup.
type Provider int64

// Provider IDs matching rows seeded in the provider_types table.
const (
	AuthentikProvider Provider = 1
)

// ID returns the database id for the provider.
func (p Provider) ID() int64 { return int64(p) }

// String returns the name stored in the provider_types table.
func (p Provider) String() string {
	switch p {
	case AuthentikProvider:
		return "authentik"
	default:
		return ""
	}
}
