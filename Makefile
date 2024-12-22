include .env
MOCKERY=./bin/mockery
SQLC_VER=1.25.0

run:
	@go run cmd/main.go

build:
	@./build.sh
	@migrate-up

mocks: install-tools
	@$(MOCKERY)

install-tools:
	@cd tools && go generate -tags=tools ./...

sqlc:
	@docker run --rm -v $(shell pwd):/src -w /src sqlc/sqlc:$(SQLC_VER) generate

migrate-create:
	@migrate create -ext sql -dir migrations/postgres tasks

migrate-up:
	@migrate -path ./migrations/postgres -database postgres://krechetov:pass@localhost:5432/tasks?sslmode=disable up

migrate-down:
	@migrate -source file://migrations/postgres -database postgres://krechetov:pass@localhost:5432/tasks?sslmode=disable down -all

.PHONY: mocks sqlc migrate-create migrate-up migrate-down
