[![Go Report Card](https://goreportcard.com/badge/github.com/bit8bytes/gearberg)](https://goreportcard.com/report/github.com/bit8bytes/gearberg)

# gearberg

Self-hostable inventory and rental management. Free as in freedom.

![mockup](/mockup/gearberg.png)

Mockup generated with [Claude Design](https://www.anthropic.com/news/claude-design-anthropic-labs) and is subject to change. This is just for inspirational purposes to get the project off and running.

## Features

**Inventory** — Organize equipment by category with descriptions, pictures, and quantities.

**Rentals** — Check items out to people, set due dates, and track partial checkouts.

Future developments will be *import* and *export* of data into CSV files. This will make it easy to switch back to paper or use any other desired provider.

For more information see current specification [here](./wiki/SPECS.md).

## Self-hosting

Self-hosting will be provided via binary for the major platforms. Container hosting will be available too. Two databases will be supported, SQLite and PostgreSQL.

## Development

We follow the [Leitfaden zur Entwicklung
sicherer Webanwendungen](https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/Studien/Webanwendungen/Webanw_Auftragnehmer.pdf?__blob=publicationFile&v=1) from the [Bundesamt für Sicherheit in der Informationstechnik (BSI)](https://www.bsi.bund.de/EN) to ensure secure software. There is no English translation yet, but we will document all steps in English. Find more info in our [wiki](./wiki/).

## License

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

Disclaimer: This license was chosen to keep this software free and self-hostable, but not for resale.