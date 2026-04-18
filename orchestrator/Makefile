.PHONY: build run test tidy

build:
	go build -o bin/orchestrator ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy
