FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY agents ./agents
WORKDIR /src/agents
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /agents ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /agents /agents
EXPOSE 8090
ENTRYPOINT ["/agents"]
