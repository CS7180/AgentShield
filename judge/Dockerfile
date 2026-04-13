FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY judge ./judge
WORKDIR /src/judge
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /judge ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /judge /judge
EXPOSE 8091
ENTRYPOINT ["/judge"]
