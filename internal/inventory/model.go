// Package inventory implements business logic and data access for inventory items.
package inventory

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrFutureInspectionDate is returned by LogInspection when InspectedAt is after today.
var ErrFutureInspectionDate = errors.New("inspection date cannot be in the future")

// PurchasePriceInput formats the purchase price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (inv *Inventory) PurchasePriceInput() string { return priceInput(inv.PurchasePrice) }

// RentalPriceInput formats the rental price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (inv *Inventory) RentalPriceInput() string { return priceInput(inv.RentalPrice) }

// priceInput formats cents as "%.2f" using dot separator. Form parsers accept both
// dot and comma, so dot is always safe as a display format for editable inputs.
func priceInput(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(*v)/100)
}

// InspectionIntervalDaysInput returns inspection interval in days as a string, or "" when nil.
func (inv *Inventory) InspectionIntervalDaysInput() string {
	return optionalInt64Input(inv.InspectionIntervalDays)
}

// WeightGInput returns weight in grams as a string, or "" when nil.
func (inv *Inventory) WeightGInput() string { return optionalInt64Input(inv.WeightG) }

// WidthMMInput returns width in mm as a string, or "" when nil.
func (inv *Inventory) WidthMMInput() string { return optionalInt64Input(inv.WidthMM) }

// HeightMMInput returns height in mm as a string, or "" when nil.
func (inv *Inventory) HeightMMInput() string { return optionalInt64Input(inv.HeightMM) }

// DepthMMInput returns depth in mm as a string, or "" when nil.
func (inv *Inventory) DepthMMInput() string { return optionalInt64Input(inv.DepthMM) }

// PowerWattsInput converts milliwatts to watts for display (e.g. 500000 → "500"). Returns "" when nil.
func (inv *Inventory) PowerWattsInput() string {
	if inv.PowerMW == nil {
		return ""
	}
	w := float64(*inv.PowerMW) / 1000
	return fmt.Sprintf("%g", w)
}

// CurrentAmpsInput converts milliamps to amps for display (e.g. 1500 → "1.5"). Returns "" when nil.
func (inv *Inventory) CurrentAmpsInput() string {
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

// Inventory represents a single inventory item.
type Inventory struct {
	ID                     string
	OrgID                  string
	Type                   Type
	UsageType              UsageType
	Name                   string
	Code                   int64
	CategoryID             string
	CategoryName           string
	CategoryColor          string
	ManufacturerID         string
	LocationID             string
	LocationName           string
	StorageObjectID        *string
	ImageURL               string
	TotalStock             int64
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

// Base holds fields shared between bulk and serialized creation.
type Base struct {
	ID             string
	OrgID          string
	UsageTypeID    int64
	Name           string
	CategoryID     string
	ManufacturerID string
	LocationID     *string
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

// CreateBulkInventory holds the data required to create a bulk inventory item.
type CreateBulkInventory struct {
	Base
	TotalStock int64
}

// CreateUnit holds the data required to create a single serialized inventory unit.
type CreateUnit struct {
	ID           string
	InventoryID  string
	SerialNumber string
}

// CreateSerializedInventory holds the data required to create a serialized inventory item with units.
type CreateSerializedInventory struct {
	Base
	Units []CreateUnit
}

// Unit represents a single serialized unit.
type Unit struct {
	ID               string
	InventoryID      string
	StatusID         int64
	UnitNumber       int64
	SerialNumber     string
	Notes            string
	PurchasedAt      *int64
	LastInspectionAt *int64 // populated from inspection records; takes precedence over PurchasedAt for scheduling
	CreatedAt        int64
	UpdatedAt        int64
}

// PurchasedAtInput formats the purchased_at timestamp as "YYYY-MM-DD" for use in
// a date input. Returns "" when nil.
func (u *Unit) PurchasedAtInput() string {
	if u.PurchasedAt == nil {
		return ""
	}
	return time.Unix(*u.PurchasedAt, 0).UTC().Format("2006-01-02")
}

// nextInspectionBase returns the unix timestamp to use as the baseline for
// scheduling the next inspection. The most recent inspection (any result) takes
// priority over purchased_at so that any logged inspection resets the clock.
func (u *Unit) nextInspectionBase() *int64 {
	if u.LastInspectionAt != nil {
		return u.LastInspectionAt
	}
	return u.PurchasedAt
}

// nextInspectionAt returns the time of the next required inspection,
// or nil when no baseline date or interval is configured.
func (u *Unit) nextInspectionAt(intervalDays *int64) *time.Time {
	base := u.nextInspectionBase()
	if base == nil || intervalDays == nil {
		return nil
	}
	t := time.Unix(*base+*intervalDays*86400, 0).UTC()
	return &t
}

// NextInspectionAtDisplay returns "YYYY-MM-DD" or "" when no baseline date or interval is available.
func (u *Unit) NextInspectionAtDisplay(intervalDays *int64) string {
	t := u.nextInspectionAt(intervalDays)
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func (u *Unit) inspectionDaysOffset(intervalDays *int64) (int64, bool) {
	next := u.nextInspectionAt(intervalDays)
	if next == nil {
		return 0, false
	}
	now := time.Now().UTC()
	// Truncate both to midnight so we compare calendar days, not clock time.
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	nextDay := time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
	return int64(nextDay.Sub(nowDay).Hours() / 24), true
}

// NextInspectionInDays returns the number of calendar days until the next
// required inspection, or "" when no baseline date or interval is available.
// Negative values are reported as "N days overdue".
func (u *Unit) NextInspectionInDays(intervalDays *int64) string {
	days, ok := u.inspectionDaysOffset(intervalDays)
	if !ok {
		return ""
	}
	switch {
	case days < 0:
		return fmt.Sprintf("%d days overdue", -days)
	case days == 0:
		return "Today"
	default:
		return fmt.Sprintf("%d days", days)
	}
}

// IsInspectionOverdue reports whether the next inspection date has passed.
func (u *Unit) IsInspectionOverdue(intervalDays *int64) bool {
	days, ok := u.inspectionDaysOffset(intervalDays)
	return ok && days < 0
}

// UpdateInventoryDetails holds the data required to update the details tab fields.
type UpdateInventoryDetails struct {
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

// UpdateInventoryPricing holds the data required to update the pricing tab fields.
type UpdateInventoryPricing struct {
	ID            string
	PurchasePrice *int64
	RentalPrice   *int64
}

// UpdateInventoryProperties holds the data required to update the properties tab fields.
type UpdateInventoryProperties struct {
	ID        string
	WeightG   *int64
	WidthMM   *int64
	HeightMM  *int64
	DepthMM   *int64
	PowerMW   *int64
	CurrentMA *int64
}

// Inspection represents a single unit inspection log entry.
type Inspection struct {
	ID          string
	UnitID      string
	InspectedAt int64
	Passed      bool
	Notes       string
	CreatedAt   int64
}

// InspectedAtDisplay formats the inspected_at timestamp as "YYYY-MM-DD".
func (i Inspection) InspectedAtDisplay() string {
	return time.Unix(i.InspectedAt, 0).UTC().Format("2006-01-02")
}

// LogInspection holds the data required to create an inspection entry.
type LogInspection struct {
	ID          string
	UnitID      string
	InspectedAt int64
	Passed      bool
	Notes       string
}

// UpdateInventoryInspection holds the data required to update the inspection tab fields.
type UpdateInventoryInspection struct {
	ID                     string
	InspectionIntervalDays *int64
}

// SetImage links or unlinks a storage object from an inventory item.
type SetImage struct {
	ID              string
	StorageObjectID *string
}

// AddUnit holds the data required to add a single unit to a serialized inventory item.
type AddUnit struct {
	ID          string
	InventoryID string
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
	ID           string
	StatusID     int64
	SerialNumber string
	Notes        string
	PurchasedAt  *int64
}
