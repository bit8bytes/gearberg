// Package equipment implements business logic and data access for inventory items.
package equipment

import (
	"fmt"
	"strings"
	"time"
)

// PurchasePriceInput formats the purchase price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (inv *Equipment) PurchasePriceInput() string { return priceInput(inv.PurchasePrice) }

// RentalPriceInput formats the rental price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (inv *Equipment) RentalPriceInput() string { return priceInput(inv.RentalPrice) }

// priceInput formats cents as "%.2f" using dot separator. Form parsers accept both
// dot and comma, so dot is always safe as a display format for editable inputs.
func priceInput(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(*v)/100)
}

// InspectionIntervalDaysInput returns inspection interval in days as a string, or "" when nil.
func (inv *Equipment) InspectionIntervalDaysInput() string {
	return optionalInt64Input(inv.InspectionIntervalDays)
}

// WeightGInput returns weight in grams as a string, or "" when nil.
func (inv *Equipment) WeightGInput() string { return optionalInt64Input(inv.WeightG) }

// WidthMMInput returns width in mm as a string, or "" when nil.
func (inv *Equipment) WidthMMInput() string { return optionalInt64Input(inv.WidthMM) }

// HeightMMInput returns height in mm as a string, or "" when nil.
func (inv *Equipment) HeightMMInput() string { return optionalInt64Input(inv.HeightMM) }

// DepthMMInput returns depth in mm as a string, or "" when nil.
func (inv *Equipment) DepthMMInput() string { return optionalInt64Input(inv.DepthMM) }

// PowerWattsInput converts milliwatts to watts for display (e.g. 500000 → "500"). Returns "" when nil.
func (inv *Equipment) PowerWattsInput() string {
	if inv.PowerMW == nil {
		return ""
	}
	w := float64(*inv.PowerMW) / 1000
	return fmt.Sprintf("%g", w)
}

// CurrentAmpsInput converts milliamps to amps for display (e.g. 1500 → "1.5"). Returns "" when nil.
func (inv *Equipment) CurrentAmpsInput() string {
	if inv.CurrentMA == nil {
		return ""
	}
	a := float64(*inv.CurrentMA) / 1000
	return fmt.Sprintf("%g", a)
}

func optionalInt64Input(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
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
	PurchasePrice          *int64
	RentalPrice            *int64
	Notes                  string
	WeightG                *int64
	WidthMM                *int64
	HeightMM               *int64
	DepthMM                *int64
	PowerMW                *int64
	CurrentMA              *int64
	InspectionIntervalDays *int64
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
	PurchasePrice  *int64
	RentalPrice    *int64
	Notes          string
	WeightG        *int64
	WidthMM        *int64
	HeightMM       *int64
	DepthMM        *int64
	PowerMW        *int64
	CurrentMA      *int64
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
	PurchasePrice            *int64
	PurchasedAt              *int64
	NextInspectionAt         *int64
	CreatedAt                int64
	UpdatedAt                int64
}

// PurchasePriceInput formats the purchase price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (u *Unit) PurchasePriceInput() string { return priceInput(u.PurchasePrice) }

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
	ID            string
	PurchasePrice *int64
	RentalPrice   *int64
}

// UpdateEquipmentProperties holds the data required to update the properties tab fields.
type UpdateEquipmentProperties struct {
	ID        string
	WeightG   *int64
	WidthMM   *int64
	HeightMM  *int64
	DepthMM   *int64
	PowerMW   *int64
	CurrentMA *int64
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
	PurchasePrice            *int64
	PurchasedAt              *int64
	NextInspectionAt         *int64
}
