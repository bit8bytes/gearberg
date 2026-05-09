# Specification (V1, 5/2/2026)

[gearberg.org](./README.md) will be a free self-hostable inventory and rental management software. The development has three milestones: Inventory, Rentals, and Import & Export. Each Milestone ends with a deliverable increment.

The first increment comes without login because the software will be self-hosted and not exposed on the internet. Login functionality will be implemented if needed later. Settings for VAT and currency will be available.

## User stories

### Settings

| ID | Name |
| - | - |
| M0.1 | A user can configure company name, currency, VAT rate, and timezone |
| M0.2 | A user can create, update, and delete inventory categories |
| M0.3 | A user cannot delete a category that is assigned to one or more inventory items |

### Milestone 1: Inventory

| ID | Name |
| - | - |
| M1.1 | A user can create, update, and delete inventory items |
| M1.2 | A user can full text search inventory by name |
| M1.3 | A user can filter inventory by category |
| M1.4 | A user can view inventory in a list |
| M1.5 | A user can see the total quantity of an inventory item |
| M1.6 | A user can view a single inventory item in detail |
| M1.7 | A user can sort inventory by name, category, stock, or price |
| M1.8 | A user can upload an image for an inventory item |
| M1.9 | A user can see how many units of an item are currently available (not rented out) |
| M1.10 | A user cannot delete an inventory item that has draft or active rental line items |

### Milestone 2: Rentals

| ID | Name |
| - | - |
| M2.1 | A user can create, update, and delete customers |
| M2.2 | A user can view customers in a list |
| M2.3 | A user can view a single customer in detail |
| M2.4 | A user can view the rental history of a customer |
| M2.5 | A user can start (create) a draft rental for an existing customer |
| M2.6 | A user can add or remove inventory items from a draft rental |
| M2.7 | A user can increase or decrease the quantity of an item in a draft rental |
| M2.8 | A user can only add an item to a rental if enough units are available -- draft rentals reserve stock |
| M2.9 | A user can set a checkout date and expected return date on a draft rental |
| M2.10 | A user can finalize a draft rental (transitions to active) |
| M2.11 | A user can mark an active rental as returned (sets return date, transitions to returned) |
| M2.12 | A user can update or delete a draft rental |
| M2.13 | A user can view all rentals in a list |
| M2.14 | A user can filter rentals by status (draft, active, returned) |
| M2.15 | A user can view overdue rentals (active rentals past their expected return date) |
| M2.16 | A user can view a single rental in detail |
| M2.17 | A user can view the cost per inventory item in a rental, charged per day (`quantity x rental_price x ceil(duration_minutes / 1440) x (1 - discount_rate)`) |
| M2.18 | A user can view the total net cost of a rental and the total including VAT (rate from company settings) as a display-only line |
| M2.19 | A user cannot delete a customer that has draft or active rentals |
| M2.20 | A user can generate an invoice for an active rental |
| M2.21 | A user can see previously generated invoices of customers on the customer detail page (see M2.3) |
| M2.22 | A user can generate a quote for a draft rental |
| M2.23 | A user can enter a discount (in %) when generating an invoice or quote |

### Milestone 3: Import & Export

| ID | Name |
| - | - |
| M3.1 | A user can import new inventory items using CSV |
| M3.2 | A user can export all inventory items using CSV |
| M3.3 | A user can export all rentals using CSV |
| M3.4 | A user can export all customers using CSV |
| M3.5 | A user can open a print view of all inventory items including images and print or export it as PDF via the browser |

### Milestone 4: Login

| ID | Name |
| - | - |
| M4.1 | A server operator can create a user with `gearberg admin create-user --email=<email> --password=<password>` |
| M4.2 | A server operator can reset any user's password with `gearberg admin reset-password --email=<email> --password=<password>` |
| M4.3 | A user can log in with email and password |
| M4.4 | A user can log out |
| M4.5 | A user can change their own password |
| M4.6 | An unauthenticated user is redirected to the login page |
| M4.7 | When started with `gearberg serve --no-auth`, the server skips authentication and all routes are accessible without login |


## Product Data

### Companies

On a self-hosted instance this table always has exactly one row, seeded at install time.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | name | Display name of the company |
| 3 | created_at | |
| 4 | updated_at | |

### Company Settings

One row per company, seeded at install time.

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | company_id | Reference to a company |
| 3 | currency | ISO 4217 currency code (e.g. `EUR`, `USD`) |
| 4 | vat_rate | VAT rate as a decimal (e.g. `0.19` for 19%). Set to `0` if not applicable. |
| 5 | timezone | IANA timezone string (e.g. `Europe/Berlin`). Used for all date calculations. |
| 6 | created_at | |
| 7 | updated_at | |

### Categories

Scoped per company. Seeded with sensible defaults at install time (e.g. Camera, Lighting, Audio, Other).

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | company_id | Reference to a company |
| 3 | name | Display name of the category |
| 4 | created_at | |
| 5 | updated_at | |

### Inventory

| ID | Name | Description |
| - | - | - |
| 1 | id | Internal identifier |
| 2 | company_id | Reference to a company |
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
| 2 | company_id | Reference to a company |
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
| 2 | company_id | Reference to a company |
| 3 | customer_id | Reference to a customer |
| 4 | status | One of 'draft', 'active', 'returned' |
| 5 | checkout_date | Date items were handed out |
| 6 | expected_return_date | Planned return date |
| 7 | return_date | Actual return date (set when marked returned) |
| 8 | notes | |
| 9 | created_at | |
| 10 | updated_at | |

### Rental Items

Line items linking an inventory item to a rental. Scoped implicitly through `rental_id` — no separate `company_id` needed.

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
