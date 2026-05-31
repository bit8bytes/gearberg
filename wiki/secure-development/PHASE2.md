# Phase 2: Conception & Planning

> **Scope:** This document covers M0–M3 (no authentication). Sections marked *deferred to M4* will be expanded when login is introduced.

## KOP 1: Description of the application

Gearberg is a self-hostable, open-source (AGPL v3) web application for inventory and rental management. Operators run it on their own infrastructure to track equipment, manage checkouts, and follow up on returns.

| Stakeholder | Goal |
|---|---|
| Operator | Self-host a private inventory system |
| Staff | Check items in/out, track overdue rentals |
| Borrower | Know what they have borrowed and when it is due |
| Administrator | Manage users, roles, and org settings |

**Functions:** inventory (items, categories, tags, availability), rentals (checkout, return, overdue tracking), multi-tenancy (isolated org data), data portability (CSV export/import), authentication (session login, RBAC — M4).

**Desired flows:** staff opens the app → records a rental with item, borrower, and due date → monitors overdue list → marks item returned → exports history to CSV.

**Undesired flows:** cross-tenant data access; manipulation of rental records; mass-import DoS; PII exposure via API or export.

| Security area | Risk |
|---|---|
| Data integrity | Tampering with rental records or inventory counts |
| Data privacy | Borrower PII exposed via API or export |
| Input validation | SQL/XSS injection via names, descriptions, CSV |
| Supply chain | Tampered binary or container image |
| Authentication *(M4)* | Credential theft, brute force, session hijacking |
| Authorization *(M4)* | Privilege escalation, tenant isolation breach |

## KOP 2: Data Handling Strategy

### Classification

| Class | Label | Definition |
|---|---|---|
| C1 | Public | May be disclosed to anyone |
| C2 | Internal | Intended for operator use only |
| C3 | Confidential | Sensitive or legally regulated |

### Data Inventory

| Data | Class | Legal basis |
|---|---|---|
| Item metadata, images, inventory counts | C1 | — |
| Org name and settings, rental records, audit log entries | C2 | — |
| Borrower identity (name, email, phone) | C3 | GDPR Art. 6 |
| User credentials, session tokens *(M4)* | C3 | GDPR Art. 6 |
| CSV exports | Inherits highest class of included fields | — |

### Handling Rules

| | C1 | C2 | C3 |
|---|---|---|---|
| Storage | Plaintext | Plaintext; DB file protected at OS level | Plaintext for PII; Argon2id for passwords *(M4)* |
| Transmission | HTTPS (HTTP on localhost only) | HTTPS | HTTPS only |
| Display | Unrestricted | Operator use only | Staff/admin only for PII; never for credentials |
| Export | Allowed | Operator use only | Allowed; `Cache-Control: no-store`, attachment |
| Logging | Allowed | Allowed; avoid full payloads | Use opaque ID for PII; never log credentials |

### Field-level Encryption

The following C3 fields should be encrypted at rest: borrower name, email, and phone number.

Field-level encryption requires a **license**. The free version stores these fields as plaintext. Operators must be informed of this explicitly so they do not assume encryption is active.

### GDPR Obligations

- **Purpose limitation** — borrower data used only for the rental relationship.
- **Data minimization** — name and contact detail only.
- **Right to erasure** — borrower PII must be anonymizable; rental event data is retained.
- **Right to access** — operator must be able to export all personal data for one person.
- **Retention** — no automated policy; operator is responsible.

## KOP 3: Roles and Permissions Concept

*Deferred to M4.* Single-operator access only in M0–M3. Target state: Admin role (user/settings management) and Staff role (inventory/rentals), addable to the same user. See SPECS.md M4.

## KOP 4: Security Requirements

### SR-INTEGRITY — Data Integrity

| ID | Requirement |
|---|---|
| SR-INTEGRITY-1 | Rental records are never hard-deleted; deletion anonymizes borrower identity, retains event data. |
| SR-INTEGRITY-2 | Mutations to rentals, inventory, and settings are written to an append-only audit log (action, resource ID, timestamp). |

### SR-PRIVACY — Data Privacy

| ID | Requirement |
|---|---|
| SR-PRIVACY-1 | GDPR erasure: anonymize all PII (name, email, phone) for a specified borrower across all records. |
| SR-PRIVACY-2 | GDPR access: export all personal data for a specified person as a single structured file. |
| SR-PRIVACY-3 | C3 CSV exports: `Content-Disposition: attachment`, `Cache-Control: no-store`. |

### SR-INPUT — Input Validation

| ID | Requirement |
|---|---|
| SR-INPUT-1 | All input validated server-side; client-side validation is additive only. |
| SR-INPUT-2 | All DB queries use parameterized statements; no raw string interpolation of user input. |
| SR-INPUT-3 | All HTML output auto-escaped via template engine; manual escaping is not the primary control. |
| SR-INPUT-4 | File uploads validated by magic bytes (JPEG, PNG, WebP); max 5 MB. |
| SR-INPUT-5 | CSV imports parsed and validated before any DB write; malformed input rejected, no partial writes. |

### SR-TRANSPORT — Transport Security

| ID | Requirement |
|---|---|
| SR-TRANSPORT-1 | `Strict-Transport-Security: max-age=31536000; includeSubDomains` on all production responses. |
| SR-TRANSPORT-2 | HTTP → HTTPS redirect in production; HTTP permitted on localhost only. |
| SR-TRANSPORT-3 | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, restrictive `Content-Security-Policy` on all HTML responses. |

### SR-SUPPLY — Supply Chain

| ID | Requirement |
|---|---|
| SR-SUPPLY-1 | Release binaries and container images are signed; signing key separate from build environment. |
| SR-SUPPLY-2 | `govulncheck` must pass clean on every CI run; known vulnerabilities block release. |
| SR-SUPPLY-3 | Release artifacts include a SHA-256 checksum file. |

### Deferred to M4

SR-AUTH (password hashing, session tokens, rate limiting, credential enumeration) and SR-AUTHZ (tenant isolation enforcement, role checks) will be added when login is introduced.

## KOP 5: Abuse Cases (M0)

### UC-M0-1 — Configure org settings (M0.1)

| ID | Abuse | Method | Impact | Mitigated by |
|---|---|---|---|---|
| AC-M0-01 | Stored XSS via org name | Submit `<script>` as name; rendered in all page headings and invoice headers | Script execution in operator's browser | SR-INPUT-3 |
| AC-M0-02 | VAT rate corruption | Submit value outside `[0.00, 1.00]`; flows into all invoice totals | Silent financial miscalculation; legal liability | SR-INPUT-1, FSR-M0-03 |
| AC-M0-03 | Invalid currency code | Submit non-ISO-4217 value; downstream formatters crash or render garbage | Broken UI; potential error exposure | SR-INPUT-1, FSR-M0-02 |
| AC-M0-04 | Malicious timezone string | Submit SQL/shell payload as timezone; used in time calculations | Corrupted duration calculations; injection surface | SR-INPUT-1, SR-INPUT-2, FSR-M0-02 |
| AC-M0-05 | Internal field overwrite | Include `id` or `created_at` in request body; handler binds full struct | Timestamp forgery in audit records | NFSR-M0-03 |
| AC-M0-06 | Error detail leakage | Send malformed payload to trigger server error; response exposes stack trace | Information disclosure | FSR-M0-05 |

### UC-M0-2/3 — Create and update a category (M0.2)

| ID | Abuse | Method | Impact | Mitigated by |
|---|---|---|---|---|
| AC-M0-07 | Stored XSS via category name | Script payload in name; rendered in inventory lists and rental selectors | Script execution in every view showing categories | SR-INPUT-3, NFSR-M0-01 |
| AC-M0-08 | SQL injection via category name | SQL metacharacters in name; query built by string concatenation | Data destruction or full DB exfiltration | SR-INPUT-2, NFSR-M0-02 |
| AC-M0-09 | Empty/whitespace-only name | Submit blank name; renders as blank in all pickers | Degraded data quality; workflow confusion | SR-INPUT-1, FSR-M0-01 |
| AC-M0-10 | Category flooding | Scripted mass-creation; no count or rate limit | DB growth; slow UI | SR-INPUT-1 |

### UC-M0-4 — Delete a category (M0.3)

| ID | Abuse | Method | Impact | Mitigated by |
|---|---|---|---|---|
| AC-M0-11 | Referential integrity bypass | Craft `DELETE /categories/{id}` directly, bypassing UI disable | Orphaned `category_id` on inventory items; silent data corruption | SR-INPUT-1, FSR-M0-04 |

## KOP 6: Threat Modeling

*Deferred to M4.* KOP 5 abuse cases cover the M0 threat surface. Full STRIDE threat model will be added when auth and multi-tenancy are introduced.

## KOP 7: Creation of the Security Architecture

*Deferred to M4.*

## KOP 8: Security Test Planning

*Deferred to M4.*

## KOP 9: Software Security Metrics

*Deferred to M4.*
