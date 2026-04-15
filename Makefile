SHELL := /bin/sh
COMPOSE_FILE := deploy/compose/docker-compose.yml

.PHONY: up down run run-api run-worker run-loadgen migrate seed test

up:
	docker compose -f $(COMPOSE_FILE) up --build

down:
	docker compose -f $(COMPOSE_FILE) down

run:
	@echo "Starting bootstrap API. Run 'make run-worker' in a second terminal for the worker placeholder."
	@$(MAKE) run-api

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

run-loadgen:
	go run ./cmd/loadgen

migrate:
	docker compose -f $(COMPOSE_FILE) run --rm api go run ./cmd/migrate

seed:
	docker compose -f $(COMPOSE_FILE) run --rm api go run ./cmd/seed

test:
	go test ./...
