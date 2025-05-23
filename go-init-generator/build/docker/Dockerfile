# Этап сборки
FROM golang:1.23-alpine AS builder
WORKDIR /service
COPY . .
RUN go build -o service ./cmd

# Этап выполнения
FROM alpine:latest
WORKDIR /service
COPY --from=builder /service/service .
COPY build/config/* .
# Копируем файлы шаблонов
COPY --from=builder /service/internal/generator/templates/microservices ../internal/generator/templates/microservices
RUN adduser -D service-runner
USER service-runner
CMD ["/service/service", "--config", "config.yml"]
