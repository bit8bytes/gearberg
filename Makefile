include .env
export

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/web: run the web application with live reload
.PHONY: run/web
run/web:
	reflex -s -r '\.(go|tmpl|js)$$' -- go run -tags sqlite ./cmd/web

## dev: start development server (web + tailwind watch)
.PHONY: dev
dev:
	$(MAKE) -j2 tailwind run/web

## lint: run linters
.PHONY: lint
lint:
	golangci-lint run --fix ./...

## license: check dependency licenses
.PHONY: license
license:
	go-licenses check ./... \
		--allowed_licenses=MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC \
		--ignore github.com/bit8bytes/gearberg \
		--ignore modernc.org/mathutil

## test: run tests
.PHONY: test
test:
	go test -shuffle=on -short -tags sqlite -vet=off -race -timeout 15s -covermode=atomic -coverprofile=/tmp/profile.out ./...

## verify: run tests, linters, and verify dependencies
.PHONY: verify
verify: lint license test
	go mod verify

## sqlc: generate source code from SQL
.PHONY: sqlc
sqlc:
	sqlc generate -f sqlc.sqlite.yml

## build: build the application
.PHONY: build
build: verify
	go build -tags sqlite -ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" -o=./bin/web ./cmd/web


## tailwind: run Tailwind CSS in watch mode
.PHONY: tailwind
tailwind:
	tailwindcss -i ./assets/css/index.css -o ./assets/dist/index.css --watch

## tailwind/build: build Tailwind CSS once (no watch)
.PHONY: tailwind/build
tailwind/build:
	tailwindcss -i ./assets/css/index.css -o ./assets/dist/index.css --minify