## lint: run linters
.PHONY: lint
lint:
	golangci-lint run --fix ./...

## license: check dependency licenses
.PHONY: license
license:
	go-licenses check ./... \
		--allowed_licenses=MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC \
		--ignore github.com/bit8bytes/filmlet \
		--ignore modernc.org/mathutil \
		--ignore github.com/segmentio/asm

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

## nix/check: check if the Nix build works
.PHONY: nix/check
nix/check:
	nix flake check --all-systems
