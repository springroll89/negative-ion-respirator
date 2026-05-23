.PHONY: dev prod test build clean

dev:
	docker compose up -d
	docker compose logs -f backend

prod:
	docker compose -f docker-compose.prod.yml up -d

down:
	docker compose down

test:
	go test ./internal/... -v -count=1

test-cover:
	go test ./internal/... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build:
	go build -o bin/server ./cmd/server

lint:
	go vet ./...
	gosec ./...

logs:
	docker compose logs -f

ps:
	docker compose ps

clean:
	docker compose down -v
	rm -f bin/server coverage.out coverage.html

deploy: build
	docker compose -f docker-compose.prod.yml up -d --build
