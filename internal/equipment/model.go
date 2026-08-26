// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package equipment implements business logic and data access for inventory items.
package equipment

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
	"github.com/bit8bytes/gearberg/internal/equipment/usage"
	"github.com/bit8bytes/gearberg/internal/units"
)

var (
	// ErrNotFound is returned when an equipment item does not exist.
	ErrNotFound = errors.New("equipment not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("equipment already exists")
	// ErrLimitExceeded is returned when an org's equipment limit is reached.
	ErrLimitExceeded = errors.New("equipment limit exceeded")
	// ErrInUse is returned when an item cannot be deleted due to active references.
	ErrInUse = errors.New("equipment is in use and cannot be deleted")
	// ErrInvalidContent is returned when a content assignment violates a domain rule.
	ErrInvalidContent = errors.New("invalid content assignment")
	// ErrNoContentTab is returned when the content tab is not enabled for an item.
	ErrNoContentTab = errors.New("content tab not enabled")
	// ErrNotSerializedUnit is returned when a unit operation is attempted on non-serialized equipment.
	ErrNotSerializedUnit = errors.New("units are only supported for serialized equipment")
	// ErrNoUnitsTab is returned when the units tab is not available for an item.
	ErrNoUnitsTab = errors.New("units tab not available for bulk equipment")
)

// Base holds fields shared between bulk and serialized creation.
type Base struct {
	OrgID          string
	EquipmentType  Type
	UsageTypeID    int64
	Name           string
	CategoryID     string
	ManufacturerID string
	LocationID     string
	Notes          string
	Pricing        Pricing
	Properties     Properties
}

// CreateEquipment holds the data required to create a new inventory item of any type.
// For Bulk, TotalStock is the initial stock quantity. For Serialized, UnitCount determines
// how many units are generated.
type CreateEquipment struct {
	Base
	ItemType   string // "bulk", "serialized", or "kit" — maps to both equipment_type_id and tracking_type_id
	TotalStock int64
	UnitCount  int64
}

// CreateBulkEquipment holds the data required to create a bulk inventory item.
type CreateBulkEquipment struct {
	ID         string
	BulkItemID string
	Base
	TotalStock int64
}

// CreateSerializedEquipment holds the data required to create a serialized inventory item with units.
type CreateSerializedEquipment struct {
	ID string
	Base
	Units []CreateUnit
}

// CreateUnit holds the data required to create a single serialized inventory unit.
type CreateUnit struct {
	ID                       string
	OrgID                    string
	EquipmentID              string
	SerialNumber             string
	ManufacturerSerialNumber string
	Remark                   string
	PurchasePrice            *units.Cents
	PurchasedAt              *int64
	NextInspectionAt         *int64
	IsActive                 bool
}

// AddUnit holds the data required to add a single unit to a serialized inventory item.
type AddUnit struct {
	ID           string
	OrgID        string
	EquipmentID  string
	SerialNumber string
}

// UpdateEquipmentDetails holds the data required to update the details tab fields.
type UpdateEquipmentDetails struct {
	ID             string
	OrgID          string
	Type           tracking.Type
	Name           string
	CategoryID     string
	ManufacturerID string
	LocationID     string
	Notes          string
	TotalStock     int64 // TotalStock is only applied for bulk inventory items.
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

// UpdateUnit holds the data required to update a single unit's editable fields.
type UpdateUnit struct {
	ID                       string
	SerialNumber             string
	IsActive                 int64
	ManufacturerSerialNumber string
	Remark                   string
	PurchasePrice            *units.Cents
	PurchasedAt              *int64
	NextInspectionAt         *int64
}

// AssignContent holds the data required to assign an equipment as content of a container.
type AssignContent struct {
	ID          string
	EquipmentID string
	MemberID    string
	Quantity    int64
}

// SetImage links or unlinks a storage object from an inventory item.
type SetImage struct {
	ID              string
	StorageObjectID *string
}

// Properties holds the physical specification fields shown on the Properties tab.
// Adding a new physical property field only requires updating this struct and PropertiesForm in form.go.
type Properties struct {
	Weight    *units.Grams
	Width     *units.Millimeters
	Height    *units.Millimeters
	Depth     *units.Millimeters
	Power     *units.Milliwatts
	Current   *units.Milliamps
	Voltage   *units.Millivolts
	WireGauge *units.WireGauge
}

// Pricing holds the price fields shown on the Pricing tab.
type Pricing struct {
	PurchasePrice *units.Cents
	RentalPrice   *units.Cents
}

// Equipment represents a single inventory item.
type Equipment struct {
	ID                     string
	OrgID                  string
	Type                   Type
	TrackingType           tracking.Type
	UsageType              usage.Type
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
	IsArchived             bool
	Notes                  string
	InspectionIntervalDays *int64
	InspectionStatus       InspectionStatus
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
	MemberType  tracking.Type
	Quantity    int64
}

// PartOf identifies a container equipment that includes this item in its content definition.
type PartOf struct {
	ID   string
	Name string
}

// Unit represents a single serialized unit.
type Unit struct {
	ID                       string
	EquipmentID              string
	StatusID                 int64
	SerialNumber             string
	ManufacturerSerialNumber string
	Remark                   string
	PurchasePrice            *units.Cents
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

// InspectionStatus is the traffic-light health state of an equipment item
// derived from its units' inspection dates.
type InspectionStatus string

const (
	// InspectionStatusNone means no inspection date is set on any unit.
	InspectionStatusNone InspectionStatus = ""
	// InspectionStatusOverdue means at least one unit's inspection date has passed.
	InspectionStatusOverdue InspectionStatus = "overdue"
	// InspectionStatusDueSoon means at least one unit's inspection is due within 30 days.
	InspectionStatusDueSoon InspectionStatus = "due-30d"
	// InspectionStatusUpToDate means all unit inspections are more than 30 days away.
	InspectionStatusUpToDate InspectionStatus = "up-to-date"
)

// NewInspectionStatus derives the status from the earliest unit inspection
// timestamp. Returns InspectionStatusNone when minNextAt is nil.
func NewInspectionStatus(minNextAt *int64) InspectionStatus {
	if minNextAt == nil {
		return InspectionStatusNone
	}
	now := time.Now().Unix()
	switch {
	case *minNextAt < now:
		return InspectionStatusOverdue
	case *minNextAt <= now+30*86400:
		return InspectionStatusDueSoon
	default:
		return InspectionStatusUpToDate
	}
}

// IsOverdue reports whether the status is overdue.
func (s InspectionStatus) IsOverdue() bool { return s == InspectionStatusOverdue }

// IsDueSoon reports whether the status is due within 30 days.
func (s InspectionStatus) IsDueSoon() bool { return s == InspectionStatusDueSoon }

// IsUpToDate reports whether all inspections are more than 30 days away.
func (s InspectionStatus) IsUpToDate() bool { return s == InspectionStatusUpToDate }

// IsActive returns true when the unit's is_active flag is set.
func (u *Unit) IsActive() bool { return u.StatusID == 1 }

// UnitStatusEntry is a row from the unit_statuses lookup table.
type UnitStatusEntry struct {
	ID   int64
	Name string
}

// DashboardStats holds aggregate inventory metrics shown on the dashboard.
type DashboardStats struct {
	// TotalValue is the sum of purchase_price × quantity for all units, in cents.
	TotalValue units.Cents
	TotalStock int64
}

// InspectionSummary is a lightweight projection used in dashboard widgets,
// joining a serialized unit with its parent equipment name.
type InspectionSummary struct {
	UnitID           string
	EquipmentID      string
	EquipmentName    string
	SerialNumber     string
	NextInspectionAt int64
}

// DaysOverdue returns how many full days past the inspection date have elapsed.
func (s *InspectionSummary) DaysOverdue() int64 {
	d := (time.Now().Unix() - s.NextInspectionAt) / 86400
	if d < 0 {
		return 0
	}
	return d
}

// DaysUntil returns how many full days remain until the inspection date.
func (s *InspectionSummary) DaysUntil() int64 {
	d := (s.NextInspectionAt - time.Now().Unix()) / 86400
	if d < 0 {
		return 0
	}
	return d
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
