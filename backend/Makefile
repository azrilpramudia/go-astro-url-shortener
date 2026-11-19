-include .env
export

BINARY_NAME := go-astro-url-shortener
MAIN_PATH   := ./cmd/web
MIGRATIONS  := db/migrations
DB_URL      ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)_test?sslmode=disable
COMPOSE     := docker compose --env-file .env -f ../docker-compose.yml

.PHONY: help run dev build test test-setup tidy \
	db-up db-down \
	migrate-create migrate-up migrate-down migrate-force \
	app-up app-down app-ps app-logs

help:
	@echo "Dev:     run dev build tidy"
	@echo "Test:    test-setup test"
	@echo "DB:      db-up db-down"
	@echo "Migrate: migrate-create name=<n> | migrate-up | migrate-down | migrate-force version=<n>"
	@echo "App:     app-up app-down app-ps app-logs"

run:
	go run $(MAIN_PATH)

dev:
	air -c .air.toml

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

test:
	go test -v -count=1 ./test/...

test-setup:
	-$(COMPOSE) exec -T postgres createdb -U $(DB_USER) $(DB_NAME)_test
	migrate -database "$(TEST_DB_URL)" -path $(MIGRATIONS) up

tidy:
	go mod tidy

db-up:
	$(COMPOSE) up -d

db-down:
	$(COMPOSE) down

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

app-up:
	$(COMPOSE) --profile app up -d --build

app-down:
	$(COMPOSE) --profile app down

app-ps:
	$(COMPOSE) --profile app ps

app-logs:
	$(COMPOSE) --profile app logs -f