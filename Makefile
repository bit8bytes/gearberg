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
	reflex -s -r '\.(go|tmpl|js)$$' -- go run -tags sqlite ./cmd/web serve

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

## fix: run go fix
.PHONY: fix
fix:
	go fix ./...

## verify: run tests, linters, and verify dependencies
.PHONY: verify
verify: fix lint license test
	go mod verify

## sqlc: generate source code from SQL
.PHONY: sqlc
sqlc:
	sqlc generate -f sqlc.sqlite.yml

## build: build the application
.PHONY: build
build:
	go build -tags sqlite -ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" -o=./bin/gearberg ./cmd/web

## build/linux-amd64: cross-compile static binary for linux/amd64
.PHONY: build/linux-amd64
build/linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags sqlite \
		-ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" \
		-o=./bin/linux/amd64/gearberg ./cmd/web

## build/linux-arm64: cross-compile static binary for linux/arm64
.PHONY: build/linux-arm64
build/linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags sqlite \
		-ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" \
		-o=./bin/linux/arm64/gearberg ./cmd/web

## build/darwin-amd64: cross-compile static binary for darwin/amd64
.PHONY: build/darwin-amd64
build/darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags sqlite \
		-ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" \
		-o=./bin/darwin/amd64/gearberg ./cmd/web

## build/darwin-arm64: cross-compile static binary for darwin/arm64
.PHONY: build/darwin-arm64
build/darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -tags sqlite \
		-ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" \
		-o=./bin/darwin/arm64/gearberg ./cmd/web

## build/windows-amd64: cross-compile static binary for windows/amd64
.PHONY: build/windows-amd64
build/windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags sqlite \
		-ldflags="-s -X main.revision=$$(git rev-parse --short HEAD)" \
		-o=./bin/windows/amd64/gearberg.exe ./cmd/web

## build/all: cross-compile static binaries for all targets
.PHONY: build/all
build/all: verify build build/linux-amd64 build/linux-arm64 build/darwin-amd64 build/darwin-arm64 build/windows-amd64

## tailwind: run Tailwind CSS in watch mode
.PHONY: tailwind
tailwind:
	tailwindcss -i ./assets/css/index.css -o ./assets/dist/index.css --watch

## tailwind/build: build Tailwind CSS once (no watch)
.PHONY: tailwind/build
tailwind/build:
	tailwindcss -i ./assets/css/index.css -o ./assets/dist/index.css --minify