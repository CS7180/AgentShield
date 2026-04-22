.PHONY: build run test test-verbose coverage coverage-check proto tidy lint docker-up docker-down orchestrator-run agents-run judge-run

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	go build -o bin/api-gateway ./cmd/server

run:
	go run ./cmd/server

orchestrator-run:
	cd ../orchestrator && go run ./cmd/server

agents-run:
	cd ../agents && go run ./cmd/server

judge-run:
	cd ../judge && go run ./cmd/server

# ── Test ──────────────────────────────────────────────────────────────────────
test:
	go test -race -cover ./...

test-verbose:
	go test -v -race -cover ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1

# Coverage gate: measures only pure-logic packages that don't require
# external infrastructure (postgres/redis/kafka/storage/orchestrator).
# Override the threshold with `COVERAGE_MIN=<n>` (default: 70).
TESTABLE_PKGS := \
	./internal/auth/... \
	./internal/domain/... \
	./internal/handler/... \
	./internal/middleware/... \
	./internal/validation/... \
	./internal/ws/...

coverage-check:
	@min=$${COVERAGE_MIN:-70}; \
	go test $(TESTABLE_PKGS) -coverprofile=coverage.out >/tmp/agentshield_apigw_cover.log 2>&1; \
	total=$$(go tool cover -func=coverage.out | awk '/^total:/{gsub("%","",$$3); print $$3}'); \
	echo "Total coverage: $${total}% (required: $${min}%)"; \
	awk "BEGIN { exit !($${total} >= $${min}) }" || (echo "coverage gate failed"; rm -f coverage.out; exit 1); \
	rm -f coverage.out

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
