up:
	docker compose -f deploy/compose/docker-compose.yml up --build

down:
	docker compose -f deploy/compose/docker-compose.yml down

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker
