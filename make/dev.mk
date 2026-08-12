## run/web/verify: check configuration and connectivity
.PHONY: run/web/verify
run/web/verify:
	go run -tags sqlite ./cmd/web verify

## run/web: run the web application with live reload
.PHONY: run/web
run/web:
	reflex -s -r '\.(go|tmpl|js)$$' -- go run -tags sqlite ./cmd/web serve \
		-log-level=debug \
		-tls-mode=local -port=443 -tls-cert-path=certs/cert.pem -tls-key-path=certs/key.pem \
		-base-url=https://localhost:443

## run/www: run the web application with live reload
.PHONY: run/www
run/www:
	reflex -s -r '\.(go|tmpl|js)$$' -- go run -tags sqlite ./cmd/www serve

## dev/web: start development server (web + tailwind watch)
.PHONY: dev/web
dev/web:
	$(MAKE) -j2 tailwind run/web

## dev/www: start development server (www + tailwind watch)
.PHONY: dev/www
dev/www:
	$(MAKE) -j2 tailwind run/www

## tailwind: run Tailwind CSS in watch mode
.PHONY: tailwind
tailwind:
	tailwindcss -i ./internal/assets/css/index.css -o ./internal/assets/dist/index.css --watch

## tailwind/build: build Tailwind CSS once (no watch)
.PHONY: tailwind/build
tailwind/build:
	tailwindcss -i ./internal/assets/css/index.css -o ./internal/assets/dist/index.css --minify

## generate/api: generate server code from OpenAPI spec
.PHONY: generate/api
generate/api:
	go generate ./api/

## sqlc: generate source code from SQL
.PHONY: sqlc
sqlc:
	sqlc generate -f sqlc.sqlite.yml