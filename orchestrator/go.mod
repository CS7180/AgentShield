module github.com/agentshield/orchestrator

go 1.24.0

require (
	github.com/agentshield/api-gateway v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.6.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.76.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/agentshield/api-gateway => ../api-gateway
