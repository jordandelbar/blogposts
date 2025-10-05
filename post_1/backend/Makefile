.PHONY: up build down test coverage-html

COVER_PKGS := $(shell go list ./... | grep -v ./tests | tr '\n' ',' | sed 's/,$$//')

up:
	docker compose up

build:
	docker compose build

down:
	docker compose down

run:
	go run ./cmd/main.go

test:
	go test ./... -v

migrate:
	migrate -path sql/schemas -database "postgres://myuser:mypassword@localhost:5432/postgres?sslmode=disable" up

unit-test:
	@mkdir -p .coverage
	go test -coverprofile=.coverage/unit.out -coverpkg=$(COVER_PKGS) $$(go list ./... | grep -v ./tests)

integration-test:
	@mkdir -p .coverage
	go test -coverprofile=.coverage/integration.out -coverpkg=$(COVER_PKGS) ./tests

coverage: unit-test integration-test
	gocovmerge .coverage/unit.out .coverage/integration.out > .coverage/combined.out
	@echo "Total coverage:"
	@go tool cover -func=.coverage/combined.out | grep total

coverage-html: coverage
	@mkdir -p .coverage
	go tool cover -html=.coverage/combined.out -o .coverage/coverage.html
	@echo "HTML coverage report generated at .coverage/coverage.html"

swagger-generate:
	swag init -g internal/infrastructure/http/docs/docs.go -o internal/infrastructure/http/docs

sqlc-generate:
	sqlc generate
