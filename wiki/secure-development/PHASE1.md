# Phase 1: Initial Planning & Procurement Phase

> **Disclaimer:** This evaluation was created with the assistance of Claude Sonnet 4.6 (Anthropic). All content has been reviewed and validated by the author.

## IPV 1: Analysis of Client Requirements

See the [specification](../SPECS.md) for more info.

## IPV 2: Identification and Assessment of Existing Security Requirements

Baseline: BSI Leitfaden zur Entwicklung sicherer Webanwendungen.

### Functional Security Requirements (FSR)

Functional requirements describe concrete, testable security behaviors derived from M0 user stories.

| ID | Category | Requirement |
| - | - | - |
| FSR-M0-01 | Data Handling | All M0 inputs (company name, currency, VAT rate, timezone, category name) MUST be validated server-side before persistence. Client-side validation is not a substitute. |
| FSR-M0-02 | Data Handling | Currency codes MUST conform to ISO 4217; timezone values MUST conform to the IANA Time Zone Database. Invalid values MUST be rejected. |
| FSR-M0-03 | Data Handling | The VAT rate MUST be validated as a decimal in [0.00, 1.00]. Values outside this range MUST be rejected. |
| FSR-M0-04 | Interfaces | The server MUST enforce the referential integrity rule of M0.3: a category with assigned inventory items MUST NOT be deletable, regardless of client-side state. |
| FSR-M0-05 | Data Handling | Internal error details (stack traces, query fragments) MUST NOT be returned to the client. Generic error messages MUST be used. |

### Non-Functional Security Requirements (NFSR)

Non-functional requirements describe the required security quality of the software. They are typically verified through reviews and security testing rather than functional tests.

| ID | Category | Requirement |
| - | - | - |
| NFSR-M0-01 | Data Handling | The application MUST be robust against stored XSS through all M0 input fields by applying context-appropriate output encoding. |
| NFSR-M0-02 | Data Handling | The application MUST be robust against SQL injection in all settings and category operations by using parameterized queries or an ORM. |
| NFSR-M0-03 | Interfaces | The API MUST expose only the minimum writable fields required. Internal fields (id, company_id, created_at, updated_at) MUST NOT be modifiable through any endpoint. |

## IPV 3: Creation of a Business Impact Analysis

gearberg is a self-hosted, single-tenant application intended for use on a private local network and not exposed to the internet. The operator is a small business (e.g., camera/audio/lighting rental). This limits both the attack surface and the blast radius of any compromise.

**1. Damage from unauthorized reading or manipulation of application data (Confidentiality / Integrity)**

Data processed by the application includes company settings, inventory items with purchase and rental prices, customer contact data (name, email, phone), rental records, and generated invoices and quotes.

- *Reading:* Exposure of customer contact data (name, email, phone) creates GDPR obligations for EU operators. Exposure of pricing data (purchase prices, rental prices) may constitute a competitive disadvantage. Severity: **low to medium**.
- *Manipulation:* Tampering with the VAT rate or currency setting would silently corrupt all downstream financial calculations (invoices, quotes). Manipulation of inventory stock or rental records could cause overbooking or financial disputes with customers. Severity: **medium**, primarily due to financial and legal consequences from incorrect invoicing.

**2. Damage from manipulation of the application itself (Integrity of the application)**

An attacker who can modify the application (e.g., via a stored XSS or a compromised server) could silently corrupt database records, generate incorrect invoices, or exfiltrate data from within the local network. Because the application is local-only and single-tenant, there is no external user impact. However, silent data corruption could go undetected for extended periods, leading to incorrect business decisions and potential legal liability from faulty billing documents. Severity: **medium**.

**3. Damage from unavailability of the application (Availability)**

| Duration | Impact |
| - | - |
| ~1 minute | Negligible — users retry. No business impact. |
| ~30 minutes | Minor disruption — staff cannot look up inventory or process checkouts; manual fallback (paper/spreadsheet) is feasible for short periods. |
| Several hours | Significant operational disruption — rental checkouts and returns cannot be processed efficiently; potential for missed or incorrectly handled rentals during peak hours. |

Since the application is self-hosted and operator-controlled, recovery is typically a service restart. There is no external SLA or dependency. Maximum damage is limited to lost operational efficiency and potential manual rework. Severity: **low to medium**.

**Overall assessment:** gearberg processes a small volume of business-relevant and personal data in a controlled, local environment. The protection needs for Confidentiality and Integrity are **medium** (driven by GDPR and financial accuracy requirements); the protection need for Availability is **low**.

## IPV 4: Creation of the Security Project Plan

Not applicable.

## IPV 5: Definition of a Software Security Group (SSG)

Not applicable.