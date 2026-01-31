# Build stage
FROM golang:1.25.5-alpine3.23 AS builder
LABEL authors="nhassl3"

WORKDIR /app

COPY . .

RUN go build -o main ./cmd/simplebank/main.go

# Run stage
FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/main .
COPY config/prod.yaml config/prod.yaml
COPY prod.env .

EXPOSE 8080
CMD ["/app/main", "--config=./config/prod.yaml", "--env=prod.env"]