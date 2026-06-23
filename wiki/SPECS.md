# Specification

[gearberg.org](./README.md) will be a free self-hostable equipment and rental management software. The development has three milestones: Equipment, Rentals, and Import & Export. Each Milestone ends with a deliverable increment.

The first increment comes without login because the software will be self-hosted and not exposed on the internet. Login functionality will be implemented if needed later. Settings for VAT and currency will be available.

## User stories

### Milestone 0: Settings

| ID | Name | Done |
| - | - | - |
| M0.1 | A user can configure org name, currency, VAT rate, and timezone | 100 % |
| M0.2 | A user can create, update, and delete equipment categories | 100 % |
| M0.3 | A user cannot delete a category that is assigned to one or more equipment items | 100 % |
| M0.4 | A user can create, update, and delete equipment manufacturers | 100 % |

### Milestone 1: Equipment

| ID | Name | Done |
| - | - | - |
| M1.1 | A user can create, update, and delete equipment items | 100 % |
| M1.2 | A user can full text search equipment by name | 100 % |
| M1.3 | A user can filter equipment by category | 100 % |
| M1.4 | A user can view equipment in a list | 100 % |
| M1.5 | A user can see the total quantity of an equipment item | 100 % |
| M1.6 | A user can view a single equipment item in detail | 100 % |
| M1.8 | A user can upload an image for an equipment item | 100 % |
| M1.9 | Each equipment unit is automatically assigned a unique 8 character serial number (e.g. ´9NAZWPYR´) | 100 % |
| M1.10 | A user can mark an equipment item as serialized, enabling individual unit tracking | 100 % |
| M1.11 | A user can add individual units to a serialized equipment item, each with a unique organization wide serial number | 100 % |
| M1.12 | A user can view all units of a serialized equipment item on its detail page | 100 % |
| M1.13 | A user can add notes to an individual unit (e.g. "broken fader") | 100 % |
| M1.14 | A user can set a next inspection date on an individual unit | 100 % |
| M1.15 | A user can view all units with an upcoming or overdue inspection date | 100 % |
| M1.16 | A user can set an operational status to active or inactive | 100 % |
| M1.17 | A user can download the QR code for a unit (serialized equipment item) as a PNG file | 100 % |
| M1.18 | A user can set weight, dimensions, power consumption (Watts), Volatage (V), and current draw (Ampere) on an equipment item | 100 % |

Deferred ideas: sorting by name, category, stock, or price (M1.7); bar codes (e.g. for trusses); periodic unit inspections; additional documents (pdf, pictures).

### Milestone 2: Import & Export

| ID | Name | Done |
| - | - | - |
| M2.1 | A user can import new equipment items using CSV | 100 % |
| M2.1 | A user can export equipment (serialized incl. units, bulk) items using CSV | 0 % |
| M2.2 | A user can open a print view of all equipment items including images and print or export it as PDF via the browser | 100 % |
| M2.3 | A user can open a print view of a rental and share or export it as PDF (Ref. M2.2) to send to customers; pricing details are shown by default | 100 % |

### Milestone 3: Login

| ID | Name |
| - | - |
| M3.1 | A user can sign up with email and password |
| M3.2 | A user can sign in with email and password |
| M3.3 | A user can sign out |
| M3.4 | An unauthenticated user is redirected to the sign in page |
| M3.5 | A user can request a password reset by entering their email; if SMTP is configured a reset link is sent, otherwise the reset token is printed to the server log |
| M3.6 | A user can reset their password by following the reset link (or entering the token from the server log) |

### Milestone 4: Rentals

| ID | Name |
| - | - |
| M4.1 | A user can see how many units of an item are currently available (not rented out) |
| M4.2 | A user can create, update, and delete customers |
| M4.3 | A user can view customers in a list |
| M4.4 | A user can view a single customer in detail |
| M4.5 | A user can view the rental history of a customer |
| M4.6 | A user can start (create) a draft rental for an existing customer, with an optional name (e.g. "BMW Munich") |
| M4.7 | A user can add or remove equipment items from a draft rental |
| M4.8 | A user can increase or decrease the quantity of an item in a draft rental |
| M4.9 | A user can only add an item to a rental if enough units are available -- draft rentals reserve stock |
| M4.10 | A user can set a checkout date and expected return date on a draft rental |
| M4.11 | A user can finalize a draft rental (transitions to active) |
| M4.12 | A user can mark an active rental as returned (sets return date, transitions to returned) |
| M4.13 | A user can update or delete a draft rental |
| M4.14 | A user can view all rentals in a list |
| M4.15 | A user can filter rentals by status (draft, active, returned) |
| M4.16 | A user can view overdue rentals (active rentals past their expected return date) |
| M4.17 | A user can view a single rental in detail |
| M4.18 | A user can view the cost per equipment item in a rental (`quantity x rental_price x billing_units x (1 - discount_rate)`), where `billing_units` is derived from `pricing_unit` and duration |
| M4.19 | A user can view the total net cost of a rental and the total including VAT (rate from org settings) as a display-only line |
| M4.20 | A user cannot delete a customer that has draft or active rentals |
| M4.21 | A user can generate an invoice for an active rental |
| M4.22 | A user can see previously generated invoices of customers on the customer detail page (see M4.4) |
| M4.23 | A user can generate a quote for a draft rental |
| M4.24 | A user can enter a discount (in %) when generating an invoice or quote |
| M4.25 | A user cannot delete an equipment item that has draft or active rental line items | 0 % |


## Product Data

### orgs

On a self-hosted instance this table always has exactly one row, seeded at install time.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | name | Display name of the org |
| 3 | created_at | |
| 4 | updated_at | |

### Org Settings

One row per org, seeded at install time.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | currency | ISO 4217 currency code (e.g. `EUR`, `USD`) |
| 4 | vat_rate | VAT rate as a decimal (e.g. `0.19` for 19%). Set to `0` if not applicable. |
| 5 | timezone | IANA timezone string (e.g. `Europe/Berlin`). Used for all date calculations. |
| 6 | created_at | |
| 7 | updated_at | |

### Categories

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | name | Display name of the category |
| 4 | color | HEX color code (e.g. `#FF5733`). Optional. Used for visual identification in the UI. |
| 5 | created_at | |
| 6 | updated_at | |

### Equipment

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | code | Auto-assigned unique numeric code, starting at `1000`. Used as human-readable identifier and base for unit codes (e.g. `1000-1`). |
| 4 | name | |
| 5 | category_id | Reference to a category |
| 6 | manufacturer | |
| 7 | image_key | Path to stored image |
| 8 | serialized | Whether individual unit tracking is enabled. Determines which stock source is used (see Bulk Stock vs Equipment Units). |
| 9 | purchase_price | Purchase price of this item (net) |
| 10 | rental_price | Rental price per unit per billing unit (net) |
| 11 | pricing_unit | Billing unit: one of `per_day`, `per_hour`, `per_week`, `flat`. Defaults to `per_day`. |
| 12 | notes | Additional notes, e.g. 'in case' |
| 13 | weight_grams | Weight in grams (optional) |
| 14 | width_mm | Width in millimeters (optional) |
| 15 | height_mm | Height in millimeters (optional) |
| 16 | depth_mm | Depth in millimeters (optional) |
| 17 | power_watts | Power consumption in Watts (optional) |
| 18 | current_ampere | Current draw in Ampere (optional) |
| 19 | created_at | |
| 20 | updated_at | |

### Bulk Stock

Quantity record for non-serialized equipment items. Only exists when `equipment.serialized = false`. One row per equipment item.

`total_stock`: total units owned.
`warehouse_stock`: **Computed** — `total_stock` minus units reserved in all draft and active rentals.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | equipment_id | Reference to parent equipment item (unique) |
| 3 | quantity | Total units owned |
| 4 | created_at | |
| 5 | updated_at | |

### Equipment Units

Individual tracked units belonging to a serialized equipment item. Only exists when `equipment.serialized = true`.

For serialized items, `total_stock = COUNT(equipment_units)` and `warehouse_stock = COUNT(equipment_units WHERE status = 'available')`.

`status`: one of `available`, `inhouse`, `damaged`, `under_repair`, `retired`.
- `available` — ready to be rented out
- `inhouse` — physically present but not available for rental (e.g. under maintenance, reserved internally)
- `damaged` — has damage, not rentable
- `under_repair` — currently being repaired
- `retired` — permanently decommissioned

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | equipment_id | Reference to parent equipment item |
| 3 | unit_number | User-visible unit label (e.g. `1`, `2`) |
| 4 | serial_number | Manufacturer serial number (optional) |
| 5 | status | Operational status (see above). Defaults to `available`. |
| 6 | notes | Free-text notes (e.g. "broken fader") |
| 7 | next_inspection_date | Date of next scheduled inspection (optional) |
| 8 | created_at | |
| 9 | updated_at | |

### Customers

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | name | Full name |
| 4 | email | |
| 5 | phone | |
| 6 | notes | |
| 7 | created_at | |
| 8 | updated_at | |

### Rentals

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | customer_id | Reference to a customer |
| 4 | name | Optional short label for the job (e.g. "BMW Munich") |
| 5 | status | One of 'draft', 'active', 'returned' |
| 6 | checkout_date | Date items were handed out |
| 7 | expected_return_date | Planned return date |
| 8 | return_date | Actual return date (set when marked returned) |
| 9 | notes | |
| 10 | created_at | |
| 11 | updated_at | |

### Rental Items

Line items linking an equipment item to a rental. Scoped implicitly through `rental_id` — no separate `org_id` needed.

`duration_minutes`: for returned rentals, derived from `checkout_date` to `return_date`. For draft and active rentals, derived from `checkout_date` to `expected_return_date`. 1 day = 1440 minutes.

`billing_units`: derived from `duration_minutes` and `pricing_unit`: `per_day` → `ceil(minutes / 1440)`, `per_hour` → `ceil(minutes / 60)`, `per_week` → `ceil(minutes / 10080)`, `flat` → `1`.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | rental_id | Reference to a rental |
| 3 | equipment_id | Reference to an equipment item |
| 4 | quantity | Number of units in this rental |
| 5 | rental_price | Snapshot of `rental_price` at the time the item was added (net) |
| 6 | pricing_unit | Snapshot of `pricing_unit` at the time the item was added |
| 7 | total_cost | **Computed:** `quantity x rental_price x billing_units` (net) |
| 8 | created_at | |
| 9 | updated_at | |
