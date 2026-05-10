## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## dev: start development server
.PHONY: dev
dev:
	go run ./cmd/app

## lint: run linters
.PHONY: lint
lint:
	golangci-lint run --fix ./...