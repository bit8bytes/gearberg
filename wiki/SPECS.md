# Specification

[gearberg.org](./README.md) will be a free self-hostable inventory and rental management software. The development has three milestones: Inventory, Rentals, and Import & Export. Each Milestone ends with a deliverable increment.

The first increment comes without login because the software will be self-hosted and not exposed on the internet. Login functionality will be implemented if needed later. Settings for VAT and currency will be available.

## User stories

### Milestone 0: Settings

| ID | Name | Done |
| - | - | - |
| M0.1 | A user can configure org name, currency, VAT rate, and timezone | 100 % |
| M0.2 | A user can create, update, and delete inventory categories | 100 % |
| M0.3 | A user cannot delete a category that is assigned to one or more inventory items | 100 % |

### Milestone 1: Inventory

| ID | Name | Done |
| - | - | - |
| M1.1 | A user can create, update, and delete inventory items | 100 % |
| M1.2 | A user can full text search inventory by name | 100 % |
| M1.3 | A user can filter inventory by category | 100 % |
| M1.4 | A user can view inventory in a list | 100 % |
| M1.5 | A user can see the total quantity of an inventory item | 100 % |
| M1.6 | A user can view a single inventory item in detail | 100 % |
| M1.7 | A user can sort inventory by name, category, stock, or price | 0 % |
| M1.8 | A user can upload an image for an inventory item | 100 % |
| M1.9 | A user can see how many units of an item are currently available (not rented out) | 0 % |
| M1.10 | A user cannot delete an inventory item that has draft or active rental line items | 0 % |
| M1.11 | A user can add a user-defined code to an inventory item (e.g. `4993`) | 100 % |
| M1.12 | A user can mark an inventory item as serialized, enabling individual unit tracking | 100 % |
| M1.13 | A user can add individual units to a serialized inventory item, each with a unit number and optional serial number | 100 % |
| M1.14 | A user can view all units of a serialized inventory item on its detail page | 100 % |
| M1.15 | A user can add notes to an individual unit (e.g. "broken fader") | 100 % |
| M1.16 | A user can set a next inspection date on an individual unit | 100 % |
| M1.17 | A user can view all units with an upcoming or overdue inspection date | 100 % |
| M1.18 | A user can set an operational status on an individual unit (available, damaged, under repair, retired) | 100 % |
| M1.19 | A user can view a QR code for an inventory item that encodes a link to its detail page, labelled with the item code | 100 % |
| M1.20 | A user can download the QR code for an inventory item as a PNG file | 100 % |

Defered ideas: inventory locations, and flight cases & groupings.

### Milestone 2: Import & Export

| ID | Name | Done |
| - | - | - |
| M2.1 | A user can import new inventory items using CSV | 100 % |
| M2.2 | A user can export all inventory items using CSV | 100 % |
| M2.3 | A user can export all rentals using CSV | 0 % |
| M2.4 | A user can export all customers using CSV | 0 % |
| M2.5 | A user can open a print view of all inventory items including images and print or export it as PDF via the browser | 100 % |

### Milestone 3: Login

| ID | Name |
| - | - |
| M3.1 | A server operator can create a user with `gearberg admin create-user --email=<email> --password=<password>` |
| M3.2 | A server operator can reset any user's password with `gearberg admin reset-password --email=<email> --password=<password>` |
| M3.3 | A user can log in with email and password |
| M3.4 | A user can log out |
| M3.5 | A user can change their own password |
| M3.6 | An unauthenticated user is redirected to the login page |
| M3.7 | When started with `gearberg serve --no-auth`, the server skips authentication and all routes are accessible without login |

### Milestone 4: Rentals

| ID | Name |
| - | - |
| M4.1 | A user can create, update, and delete customers |
| M4.2 | A user can view customers in a list |
| M4.3 | A user can view a single customer in detail |
| M4.4 | A user can view the rental history of a customer |
| M4.5 | A user can start (create) a draft rental for an existing customer |
| M4.6 | A user can add or remove inventory items from a draft rental |
| M4.7 | A user can increase or decrease the quantity of an item in a draft rental |
| M4.8 | A user can only add an item to a rental if enough units are available -- draft rentals reserve stock |
| M4.9 | A user can set a checkout date and expected return date on a draft rental |
| M4.10 | A user can finalize a draft rental (transitions to active) |
| M4.11 | A user can mark an active rental as returned (sets return date, transitions to returned) |
| M4.12 | A user can update or delete a draft rental |
| M4.13 | A user can view all rentals in a list |
| M4.14 | A user can filter rentals by status (draft, active, returned) |
| M4.15 | A user can view overdue rentals (active rentals past their expected return date) |
| M4.16 | A user can view a single rental in detail |
| M4.17 | A user can view the cost per inventory item in a rental, charged per day (`quantity x rental_price x ceil(duration_minutes / 1440) x (1 - discount_rate)`) |
| M4.18 | A user can view the total net cost of a rental and the total including VAT (rate from org settings) as a display-only line |
| M4.19 | A user cannot delete a customer that has draft or active rentals |
| M4.20 | A user can generate an invoice for an active rental |
| M4.21 | A user can see previously generated invoices of customers on the customer detail page (see M4.3) |
| M4.22 | A user can generate a quote for a draft rental |
| M4.23 | A user can enter a discount (in %) when generating an invoice or quote |


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

Scoped per org. Seeded with sensible defaults at install time (e.g. Camera, Lighting, Audio, Other).

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | name | Display name of the category |
| 4 | created_at | |
| 5 | updated_at | |

### Inventory

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | org_id | Reference to a org |
| 3 | name | |
| 4 | category_id | Reference to a category |
| 5 | manufacturer | |
| 6 | image_key | Path to stored image |
| 7 | total_stock | Total units owned |
| 8 | warehouse_stock | **Computed:** `total_stock` minus units in all draft and active rentals |
| 9 | purchase_price | Purchase price of this item (net) |
| 10 | rental_price | Rental price per unit per day (net) |
| 11 | notes | Additional notes, e.g. 'in case' |
| 12 | created_at | |
| 13 | updated_at | |

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
| 4 | status | One of 'draft', 'active', 'returned' |
| 5 | checkout_date | Date items were handed out |
| 6 | expected_return_date | Planned return date |
| 7 | return_date | Actual return date (set when marked returned) |
| 8 | notes | |
| 9 | created_at | |
| 10 | updated_at | |

### Rental Items

Line items linking an inventory item to a rental. Scoped implicitly through `rental_id` — no separate `org_id` needed.

`duration_minutes`: for returned rentals, derived from `checkout_date` to `return_date`. For draft and active rentals, derived from `checkout_date` to `expected_return_date`. 1 day = 1440 minutes.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | rental_id | Reference to a rental |
| 3 | inventory_id | Reference to an inventory item |
| 4 | quantity | Number of units in this rental |
| 5 | rental_price | Snapshot of the rental price per day at the time the item was added (net) |
| 6 | total_cost | **Computed:** `quantity x rental_price x ceil(duration_minutes / 1440)` (net) |
| 7 | created_at | |
| 8 | updated_at | |
