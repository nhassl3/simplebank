# Build stage
FROM golang:1.25.5-alpine3.23 AS builder
LABEL authors="nhassl3"

WORKDIR /app

COPY . .

# Build simplebank latest version
RUN go build -o main ./cmd/simplebank/main.go

# Install curl
RUN apk add curl

# Install migrate version 4.19.1
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY internals/db/migration ./migration
COPY config/prod.yaml config/prod.yaml
COPY start.sh .
COPY prod.env .

EXPOSE 8080
CMD ["/app/main", "--config=./config/prod.yaml", "--env=prod.env"]
ENTRYPOINT [ "/app/start.sh" ]