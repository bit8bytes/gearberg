package orgs

// Role represents an organization role (Owner, Admin, Member, Viewer).
// The integer value matches the id stored in the org_roles table,
// which is seeded at startup rather than via migrations.
type Role int64

// Role IDs matching rows seeded in the org_roles table.
const (
	OwnerRole  Role = 1
	AdminRole  Role = 2
	MemberRole Role = 3
	ViewerRole Role = 4
)

// ID returns the database id for the role.
func (r Role) ID() int64 { return int64(r) }

// String returns the name stored in the org_roles table.
func (r Role) String() string {
	switch r {
	case OwnerRole:
		return "Owner"
	case AdminRole:
		return "Admin"
	case MemberRole:
		return "Member"
	case ViewerRole:
		return "Viewer"
	default:
		return ""
	}
}

// Rank returns the numeric rank for the role used in permission comparisons.
func (r Role) Rank() int64 {
	switch r {
	case OwnerRole:
		return 100
	case AdminRole:
		return 50
	case MemberRole:
		return 25
	case ViewerRole:
		return 0
	default:
		return 0
	}
}
