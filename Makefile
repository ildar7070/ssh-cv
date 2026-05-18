.PHONY: run build tidy fmt vet test docker up down logs ssh

# Local Go (requires Go 1.23+).
run:
	go run ./cmd/i12k

build:
	go build -o bin/i12k ./cmd/i12k

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

# Docker / compose.
docker:
	docker build -t i12k .

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

# Convenience: connect to the running container.
ssh:
	ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
