package credentials

// Type represents a credential type. The integer value matches the id stored
// in the credential_types table, which is seeded at startup.
type Type int64

// Credential type IDs matching rows seeded in the credential_types table.
const (
	PasswordType Type = 1
)

// ID returns the database id for the credential type.
func (t Type) ID() int64 { return int64(t) }

// String returns the name stored in the credential_types table.
func (t Type) String() string {
	switch t {
	case PasswordType:
		return "password"
	default:
		return ""
	}
}
