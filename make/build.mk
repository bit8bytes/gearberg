## nix/build/web: build the Go binary with Nix
.PHONY: nix/build/web
nix/build/web:
	nix build

## nix/build/www: build the Go binary with Nix
.PHONY: nix/build/www
nix/build/www:
	nix build .#www

## nix/build/docs: build the Go binary with Nix
.PHONY: nix/build/docs
nix/build/docs:
	nix build .#docs

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