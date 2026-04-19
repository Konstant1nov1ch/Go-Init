# Контекст сборки: корень репозитория (Go-Init), см. docker-compose-all.yml.
# Нужен доступ к ../go-init-common из go.mod replace.

FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go-init-common /src/go-init-common
COPY go-init-manager /src/go-init-manager
WORKDIR /src/go-init-manager
RUN go mod download && go mod tidy && CGO_ENABLED=0 go build -o /tmp/service ./cmd

FROM alpine:latest
WORKDIR /service
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /tmp/service /service/service
COPY go-init-manager/build/config/ /service/
RUN adduser -D service-runner
USER service-runner
CMD ["/service/service", "--config", "config.yml"]
