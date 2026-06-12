# Gearberg

Self-hostable inventory and rental management. Free as in freedom.

![mockup](/mockup/gearberg.png)

Mockup generated with [Claude Design](https://www.anthropic.com/news/claude-design-anthropic-labs) and is subject to change. This is just for inspirational purposes to get the project off and running.

## Features

**Inventory** — Organize equipment by category with descriptions, pictures, and quantities.

**Rentals** — Check items out to people, set due dates, and track partial checkouts.

Future developments will be *import* and *export* of data into CSV files. This will make it easy to switch back to paper or use any other desired provider.

For more information see current specification [here](./wiki/SPECS.md).

## Self-hosting: Quickstart

Try it with a single [Docker](https://docker.com) command:

```sh
docker run -it -p 8080:8080 nixos/nix \
  nix --extra-experimental-features "nix-command flakes" \
  run github:bit8bytes/gearberg -- serve
```

Then open `http://localhost:8080` in your browser.

Self-hosting is provided via NixOS with a full NixOS configuration included.

## Contributing

**Obtain** — Download the latest release from the [releases page](https://github.com/bit8bytes/gearberg/releases), or run it directly with Nix as shown above.

**Feedback** — Open a [GitHub issue](https://github.com/bit8bytes/gearberg/issues) to report bugs or request features. Please include steps to reproduce for bug reports.

**Contribute** — See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to get involved. The project is in early development; reach out via an issue before submitting code.

## Development

[![Go Report Card](https://goreportcard.com/badge/github.com/bit8bytes/gearberg)](https://goreportcard.com/report/github.com/bit8bytes/gearberg)
[![REUSE status](https://api.reuse.software/badge/github.com/bit8bytes/gearberg)](https://api.reuse.software/info/github.com/bit8bytes/gearberg)

We follow the [Leitfaden zur Entwicklung
sicherer Webanwendungen](https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/Studien/Webanwendungen/Webanw_Auftragnehmer.pdf?__blob=publicationFile&v=1) from the [Bundesamt für Sicherheit in der Informationstechnik (BSI)](https://www.bsi.bund.de/EN) to ensure secure software. There is no English translation yet, but we will document all steps in English. Find more info in our [wiki](./wiki/).

## License

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

Disclaimer: This license was chosen to keep this software free and self-hostable, but not for resale.