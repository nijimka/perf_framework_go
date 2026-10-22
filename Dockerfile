ARG CACHE_BUST=1

FROM golang:1.27.2-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o perf-runner ./cmd/runner

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/perf-runner .
COPY config.yaml .

ENTRYPOINT ["./perf-runner"]
CMD ["-config", "config.yaml"]
