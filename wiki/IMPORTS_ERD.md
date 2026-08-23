# Database: Format agnostic import process

The goal is to let the user bring any format (start with csv), upload it, map it to Gearberg internal, and import it into the system.

This requires a bigger refactor, but we are in alpha, so we can completly delete the current design and make this possible.

We will gain easy of use and better adoption by designing a better import process.

```mermaid
erDiagram
  import_sessions {
    text id PK
    text org_id FK
    text format "csv | json | excel"
    text status "uploading | mapping | staged | committed"
    text target_entity "equipment"
    integer created_at "NOT NULL DEFAULT unixepoch()"
  }

  import_data {
    text id PK
    text session_id "REFERENCES import_sessions(id) ON DELETE CASCADE"
    integer row_number "NOT NULL CHECK row_number > 0 UNIQUE session_id+row_number"
    text data "NOT NULL DEFAULT {} JSON blob of raw source columns"
    text status "NOT NULL DEFAULT new | error | needs_review"
    text error_message "NOT NULL DEFAULT empty"
    text action "NOT NULL DEFAULT pending | create | skip"
  }

  import_mappings {
    text id PK
    text session_id "REFERENCES import_sessions(id) ON DELETE CASCADE"
    text source_col "NOT NULL UNIQUE per session e.g. Bezeichnung"
    text target_field "NOT NULL e.g. name"
  }

  import_sessions ||--o{ import_data : "has rows"
  import_sessions ||--o{ import_mappings : "has mappings"
```

The flow consists of following steps:

1. Upload CSV
2. Create import session
3. Populate `import_data` with raw rows
4. User maps source columns to Gearberg fields (`import_mappings`)
5. System validates and applies mappings, runs validation steps, marks rows
6. User reviews and resolves decision
7. User confirms and commit to inventory/equipment

We don't support editing in the UI yet, but we can do this easily by adding an override field to the import data table.

This flow can be put simply into:

1. Stage (NewSession)
2. Review
3. Commit

to abstract the complex logic behind a simple interface.

