.PHONY: run build tidy fmt vet test ci docker up down logs ssh

# Local Go (requires Go 1.23+).
run:
	go run ./cmd/ssh-cv

build:
	go build -o bin/ssh-cv ./cmd/ssh-cv

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

ci: fmt vet test build

# Docker / compose.
docker:
	docker build -t ssh-cv .

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

# Convenience: connect to the running container.
ssh:
	ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
