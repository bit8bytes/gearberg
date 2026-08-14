# Gearberg

Equipment rental software that doesn't rent itself to you. Manage your gear inventory, track rentals, handle customers, and generate invoices. Self-hostable and free.

![mockup](/mockup/gearberg-demo.gif)

## Features

**Equipment**: Bulk quantities or serialized items, with categories and photos.

**Rentals**: Check out gear to people, set due dates, track returns.

**Import/Export**: CSV in, CSV out. Your data is never locked in.

For more information see current specification [here](./wiki/SPECS.md).

## Self-hosting: Quickstart

Try it with a single [Docker](https://docker.com) command:

```sh
docker run -p 8080:8080 ghcr.io/bit8bytes/gearberg serve
```

Then open `http://localhost:8080` in your browser.

## Development

We follow the [Leitfaden zur Entwicklung
sicherer Webanwendungen](https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/Studien/Webanwendungen/Webanw_Auftragnehmer.pdf?__blob=publicationFile&v=1) from the [Bundesamt für Sicherheit in der Informationstechnik (BSI)](https://www.bsi.bund.de/EN) to ensure secure software. There is no English translation yet, but we will document all steps in English. Find more info in our [wiki](./wiki/).

## License

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

Disclaimer: This license was chosen to keep this software free and self-hostable, but not for resale.