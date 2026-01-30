BINARY_NAME=simplebank
BUILD_DIR=build

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

DATABASE_PASSWORD ?= $$DATABASE_PASSWORD

createdb:
	docker exec -it postgres18 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres18 dropdb simple_bank

opendb:
	docker exec -it postgres18 psql -U root simple_bank

postgres:
	docker run --name postgres18 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=$(DATABASE_PASSWORD) -d postgres:18-alpine

migrate-up:
	migrate -path internals/db/migration -database "postgres://root:$(DATABASE_PASSWORD)@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrate-down:
	migrate -path internals/db/migration -database "postgres://root:$(DATABASE_PASSWORD)@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrate-up-once:
	migrate -path internals/db/migration -database "postgres://root:$(DATABASE_PASSWORD)@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migrate-down-once:
	migrate -path internals/db/migration -database "postgres://root:$(DATABASE_PASSWORD)@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

mock:
	mockgen -package=mockdb -destination=./internals/db/mock/store.go github.com/nhassl3/simplebank/internals/db/sqlc Store


GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0
LDFLAGS ?= -s -w
BUILD_TAGS ?= ""

# Dynamic build based on environment
build:
	@echo "Building with: GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED)"
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build \
		-tags $(BUILD_TAGS) \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) \
		./cmd/$(BINARY_NAME)

run: build
	@echo "\n"
	@./$(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) --config=./config/local.yaml --env=.env

clean:
	rm -rf ./$(BUILD_DIR)/

test:
	go test -v -cover ./...

.PHONY: createdb dropdb opendb postgres migrate-up migrate-down sqlc build run clean test mock migrate-down-once migrate-up-once
