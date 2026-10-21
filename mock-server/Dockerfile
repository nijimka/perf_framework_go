FROM golang:1.27.2-alpine AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o mock-server main.go

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/mock-server .
EXPOSE 8080
CMD ["./mock-server"]
