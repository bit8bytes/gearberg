# Specification Revision (V2, 5/30/2026)

## Milestone 1: Inventory (additions)

| ID | Name | Done |
| - | - | - |
| M1.11 | A user can add a user-defined code to an inventory item (e.g. `4993`) | 0 % |
| M1.12 | A user can mark an inventory item as serialized, enabling individual unit tracking | 0 % |
| M1.13 | A user can add individual units to a serialized inventory item, each with a unit number and optional serial number | 0 % |
| M1.14 | A user can view all units of a serialized inventory item on its detail page | 0 % |
| M1.15 | A user can add notes to an individual unit (e.g. "broken fader") | 0 % |
| M1.16 | A user can set a next inspection date on an individual unit | 0 % |
| M1.17 | A user can view all units with an upcoming or overdue inspection date | 0 % |
| M1.18 | A user can set an operational status on an individual unit (available, in service, retired) | 0 % |

## Deferred Ideas

Not in scope for current milestones but worth revisiting:

- **QR code generation & scanning**: Generate a QR code sticker per serialized unit. Scan on checkout/checkin to confirm the exact unit going out. Requires M1.12–M1.13 to be done first.
- **Flight cases / groupings**: Group related equipment into a named kit or flight case (e.g. "X32 FOH Kit" containing the console, cables, and rack). Useful for recurring setups that always go out together. Could be implemented after M1.13.