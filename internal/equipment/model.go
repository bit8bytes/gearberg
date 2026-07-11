// Package equipment implements business logic and data access for inventory items.
package equipment

import (
	"fmt"
	"strings"
	"time"
)

// Cents is a monetary amount in the smallest currency unit (e.g. 1999 = €19.99).
type Cents int64

// ToDecimal formats the amount as a decimal string for form inputs (e.g. 1999 → "19.99").
// Returns "" when nil.
func (c *Cents) ToDecimal() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(*c)/100)
}

// Grams is a weight stored in grams; users enter and see kilograms.
type Grams int64

// ToKG returns the weight in kg as a string for form inputs (e.g. 1500 → "1.5"). Returns "" when nil.
func (g *Grams) ToKG() string {
	if g == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*g)/1000)
}

// Millimeters is a length stored in millimetres; users enter and see centimetres.
type Millimeters int64

// ToCM returns the length in cm as a string for form inputs (e.g. 100 → "10"). Returns "" when nil.
func (m *Millimeters) ToCM() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*m)/10)
}

// Milliwatts is power stored in milliwatts; users enter and see watts.
type Milliwatts int64

// ToW returns the power in W as a string for form inputs (e.g. 1500 → "1.5"). Returns "" when nil.
func (m *Milliwatts) ToW() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*m)/1000)
}

// Milliamps is current stored in milliamps; users enter and see amps.
type Milliamps int64

// ToA returns the current in A as a string for form inputs (e.g. 500 → "0.5"). Returns "" when nil.
func (m *Milliamps) ToA() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*m)/1000)
}

// Millivolts is voltage stored and displayed as whole volts.
type Millivolts int64

// ToV returns the voltage as a string for form inputs. Returns "" when nil.
func (v *Millivolts) ToV() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*v)/1000)
}

// WireGauge stores wire cross-section area as mm²×100 (e.g. 150 = 1.5 mm²); displayed as-is.
type WireGauge int64

// String returns the raw mm²×100 value as a string for form inputs. Returns "" when nil.
func (w *WireGauge) String() string {
	if w == nil {
		return ""
	}
	return fmt.Sprintf("%d", *w)
}

// Properties holds the physical specification fields shown on the Properties tab.
// Adding a new physical property field only requires updating this struct,
// the mapper functions in mapping.go, and PropertiesForm in form.go.
type Properties struct {
	Weight    *Grams
	Width     *Millimeters
	Height    *Millimeters
	Depth     *Millimeters
	Power     *Milliwatts
	Current   *Milliamps
	Voltage   *Millivolts
	WireGauge *WireGauge
}

// Pricing holds the price fields shown on the Pricing tab.
type Pricing struct {
	PurchasePrice *Cents
	RentalPrice   *Cents
}

// Equipment represents a single inventory item.
type Equipment struct {
	ID                     string
	OrgID                  string
	Kind                   Kind
	Type                   Type
	UsageType              UsageType
	Name                   string
	CategoryID             string
	CategoryName           string
	ManufacturerID         string
	LocationID             string
	LocationName           string
	StorageObjectID        *string
	ImageURL               string
	TotalStock             int64
	ContentCount           int64
	HasContent             bool
	IsArchived             bool
	Notes                  string
	InspectionIntervalDays *int64
	Pricing                Pricing
	Properties             Properties
	CreatedAt              int64
	UpdatedAt              int64
}

// ContentItem represents a single entry in an equipment's content definition.
type ContentItem struct {
	ID          string
	EquipmentID string
	MemberID    string
	MemberName  string
	MemberType  Type
	Quantity    int64
}

// AssignContent holds the data required to assign an equipment as content of a container.
type AssignContent struct {
	ID          string
	EquipmentID string
	MemberID    string
	Quantity    int64
}

// RemoveContent holds the data required to remove a content entry.
type RemoveContent struct {
	ID string
}

// PartOf identifies a container equipment that includes this item in its content definition.
type PartOf struct {
	ID   string
	Name string
}

// Base holds fields shared between bulk and serialized creation.
type Base struct {
	ID             string
	OrgID          string
	UsageTypeID    int64
	Name           string
	CategoryID     string
	ManufacturerID string
	LocationID     *string
	HasContent     bool
	Notes          string
	Pricing        Pricing
	Properties     Properties
}

// CreateBulkEquipment holds the data required to create a bulk inventory item.
type CreateBulkEquipment struct {
	Base
	TotalStock int64
}

// CreateUnit holds the data required to create a single serialized inventory unit.
type CreateUnit struct {
	ID           string
	OrgID        string
	EquipmentID  string
	SerialNumber string
	Code         string
}

// CreateSerializedEquipment holds the data required to create a serialized inventory item with units.
type CreateSerializedEquipment struct {
	Base
	Units []CreateUnit
}

// CreateEquipment holds the data required to create a new inventory item of any type.
// For Bulk, TotalStock is the initial stock quantity. For Serialized, UnitCount determines
// how many units are generated.
type CreateEquipment struct {
	Type Type
	Base
	TotalStock int64
	UnitCount  int64
}

// Unit represents a single serialized unit.
type Unit struct {
	ID                       string
	EquipmentID              string
	StatusID                 int64
	SerialNumber             string
	Code                     string
	ManufacturerSerialNumber string
	Notes                    string
	PurchasePrice            *Cents
	PurchasedAt              *int64
	NextInspectionAt         *int64
	CreatedAt                int64
	UpdatedAt                int64
}

// PurchasedAtInput formats the purchased_at timestamp as "YYYY-MM-DD" for use in
// a date input. Returns "" when nil.
func (u *Unit) PurchasedAtInput() string {
	if u.PurchasedAt == nil {
		return ""
	}
	return time.Unix(*u.PurchasedAt, 0).UTC().Format("2006-01-02")
}

// NextInspectionAtInput formats the next_inspection_at timestamp as "YYYY-MM-DD". Returns "" when nil.
func (u *Unit) NextInspectionAtInput() string {
	if u.NextInspectionAt == nil {
		return ""
	}
	return time.Unix(*u.NextInspectionAt, 0).UTC().Format("2006-01-02")
}

// NextInspectionLabel returns a human-readable countdown: "Xd overdue", "in Xd", or "" when nil.
func (u *Unit) NextInspectionLabel() string {
	if u.NextInspectionAt == nil {
		return ""
	}
	days := (*u.NextInspectionAt - time.Now().Unix()) / 86400
	if days < 0 {
		return fmt.Sprintf("%dd overdue", -days)
	}
	return fmt.Sprintf("in %dd", days)
}

// IsInspectionOverdue returns true when the next inspection date has passed.
func (u *Unit) IsInspectionOverdue() bool {
	return u.NextInspectionAt != nil && *u.NextInspectionAt < time.Now().Unix()
}

// IsActive returns true when the unit's is_active flag is set.
func (u *Unit) IsActive() bool { return u.StatusID == 1 }

// UpdateEquipmentDetails holds the data required to update the details tab fields.
type UpdateEquipmentDetails struct {
	ID             string
	Type           Type
	Name           string
	CategoryID     string
	ManufacturerID string
	LocationID     string
	Notes          string
	// TotalStock is only applied for bulk inventory items.
	TotalStock int64
}

// UpdateEquipmentPricing holds the data required to update the pricing tab fields.
type UpdateEquipmentPricing struct {
	ID      string
	Pricing Pricing
}

// UpdateEquipmentProperties holds the data required to update the properties tab fields.
type UpdateEquipmentProperties struct {
	ID         string
	Properties Properties
}

// ArchiveEquipment holds the data required to archive or unarchive an equipment item.
type ArchiveEquipment struct {
	ID         string
	IsArchived bool
}

// SetImage links or unlinks a storage object from an inventory item.
type SetImage struct {
	ID              string
	StorageObjectID *string
}

// AddUnit holds the data required to add a single unit to a serialized inventory item.
type AddUnit struct {
	ID           string
	OrgID        string
	EquipmentID  string
	SerialNumber string
	Code         string
}

// UnitStatusEntry is a row from the unit_statuses lookup table.
type UnitStatusEntry struct {
	ID   int64
	Name string
}

// Label converts the snake_case Name to Title Case (e.g. "under_repair" → "Under Repair").
func (u UnitStatusEntry) Label() string {
	words := strings.Fields(strings.ReplaceAll(u.Name, "_", " "))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// UpdateUnit holds the data required to update a single unit's editable fields.
type UpdateUnit struct {
	ID                       string
	StatusID                 int64
	Code                     string
	ManufacturerSerialNumber string
	Notes                    string
	PurchasePrice            *Cents
	PurchasedAt              *int64
	NextInspectionAt         *int64
}
