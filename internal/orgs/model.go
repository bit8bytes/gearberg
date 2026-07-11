package orgs

// Org represents a org entity.
type Org struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}
