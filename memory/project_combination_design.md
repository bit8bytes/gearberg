---
name: Virtual vs Physical Combination Design
description: Agreed architecture separating virtual combinations (concept-level recipe) from physical combinations (unit-level assignment)
type: project
---

Virtual and physical combinations are separate concerns and must not be mixed.

**Virtual combination** (`equipment_type = virtual`):
- Managed via `equipment_combination_items` — concept-level recipe ("a Case contains a Mixer")
- No `equipment_items` rows (no units)
- Many-to-many valid: a Mixer concept can appear in multiple virtual combinations
- `PartOf` in the UI applies here — shows concept-level membership

**Physical combination** (`equipment_type = physical`, `tracking_type = serialized`):
- Unit assignment via `equipment_items.parent_equipment_item_id` — "Case Unit #1 contains Mixer Unit #3"
- `has_content` toggle is removed from the physical serialized creation flow
- A physical serialized item implicitly becomes a container when units are assigned into it
- The grouping tab for a physical serialized unit shows/manages which units are inside it
- `PartOf` in the header does NOT apply to physical items (unit assignment is shown differently)

**Why:** The current `has_content` + `equipment_combination_items` approach conflates concept-level recipes with unit-level tracking. Keeping them separate makes each model clean and unambiguous.

**How to apply:** `equipment_combination_items` exclusively belongs to virtual combinations. Physical container relationships live in `equipment_items.parent_equipment_item_id` only.
