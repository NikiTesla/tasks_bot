include .env
MOCKERY=./bin/mockery

run:
	@go run cmd/main.go

build:
	@./build.sh

mocks: install-tools
	@$(MOCKERY)

install-tools:
	@cd tools && go generate -tags=tools ./...

.PHONY: mocks