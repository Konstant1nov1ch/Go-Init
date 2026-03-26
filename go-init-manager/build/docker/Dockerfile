# Этап сборки
FROM golang:1.23-alpine AS builder
WORKDIR /service
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go mod tidy && go build -o service ./cmd

# Этап выполнения
FROM alpine:latest
WORKDIR /service
COPY --from=builder /service/service .
COPY build/config/* .
RUN adduser -D service-runner
USER service-runner
CMD ["/service/service", "--config", "config.yml"]
