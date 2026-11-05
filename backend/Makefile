-include .env
export

APP_NAME   := go-astro-url-shortener
MAIN_PATH  := ./cmd/web
MIGRATIONS := db/migrations
DB_URL     ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)_test?sslmode=disable

.PHONY: help run dev build test tidy db-up db-down migrate-create migrate-up migrate-down migrate-force

help:
	@echo "Targets: run dev build test tidy db-up db-down migrate-create migrate-up migrate-down migrate-force"

run:
	go run $(MAIN_PATH)

dev:
	air

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

test:
	go test -v -count=1 ./test/...

test-setup:
	-docker compose --env-file .env -f ../docker-compose.yml exec -T postgres createdb -U $(DB_USER) $(DB_NAME)_test
	migrate -database "$(TEST_DB_URL)" -path $(MIGRATIONS) up

tidy:
	go mod tidy

db-up:
	docker compose --env-file .env -f ../docker-compose.yml up -d

db-down:
	docker compose --env-file .env -f ../docker-compose.yml down

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<migration_name>" && exit 1)
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS) up

migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS) down 1

migrate-force:
	@test -n "$(version)" || (echo "usage: make migrate-force version=<n>" && exit 1)
	migrate -database "$(DB_URL)" -path $(MIGRATIONS) force $(version)