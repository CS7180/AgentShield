.PHONY: build run test proto tidy lint docker-up docker-down

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	go build -o bin/api-gateway ./cmd/server

run:
	go run ./cmd/server

# ── Test ──────────────────────────────────────────────────────────────────────
test:
	go test -race -cover ./...

test-verbose:
	go test -v -race -cover ./...

# ── Dependencies ──────────────────────────────────────────────────────────────
tidy:
	go mod tidy

# ── Proto codegen ─────────────────────────────────────────────────────────────
# Prerequisites: protoc, protoc-gen-go, protoc-gen-go-grpc
# Install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#          go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/orchestrator/orchestrator.proto
	@echo "Proto files generated. Replace manual stubs in proto/orchestrator/"

# ── Lint ──────────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ── Docker ────────────────────────────────────────────────────────────────────
docker-up:
	docker compose up -d redis kafka zookeeper

docker-down:
	docker compose down

docker-build:
	docker compose build api-gateway

docker-full:
	docker compose up -d
