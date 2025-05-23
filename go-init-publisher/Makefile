.PHONY: build run clean logs shell help test lint format bin-deps protoc

LOCAL_BIN := $(CURDIR)/bin

## Установка необходимых плагинов для генерации protobuf
bin-deps:
	@echo "Installing protobuf plugins..."
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

## Генерация protobuf/gRPC кода
protoc: bin-deps
	@echo "Generating protobuf for archive_publisher.proto..."
	protoc \
		--experimental_allow_proto3_optional=true \
		-I api/grpc \
		--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
		--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
		--go_out pkg/api/grpc \
		--go_opt paths=source_relative \
		--go-grpc_out pkg/api/grpc \
		--go-grpc_opt paths=source_relative \
		api/grpc/archive_publisher.proto

	@echo "Running go mod tidy..."
	go mod tidy

# Загружаем переменные из .env (если он есть)
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Если переменные не заданы в .env, устанавливаем значения по умолчанию
SERVICE_NAME ?= my-service
DOCKERFILE ?= build/docker/Dockerfile
BUILD_CONTEXT ?= .

# Определяем текущую версию Go из go.mod
GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')

# ======== Docker Build ========
build:
	@echo "Building Docker image for $(SERVICE_NAME) with Go $(GO_VERSION)..."
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(SERVICE_NAME) -f $(DOCKERFILE) $(BUILD_CONTEXT)

run:
	@echo "Running $(SERVICE_NAME)..."
	docker run --rm -p 8080:8080 $(SERVICE_NAME)

clean:
	@echo "Cleaning up Docker cache..."
	docker rmi $(SERVICE_NAME) || true
	docker system prune -f

logs:
	@echo "Fetching logs from $(SERVICE_NAME)..."
	docker logs -f $(SERVICE_NAME)

shell:
	@echo "Opening shell in $(SERVICE_NAME)..."
	docker run --rm -it $(SERVICE_NAME) sh

help:
	@echo "Available make commands:"
	@echo "  make build          - Собрать Docker-образ"
	@echo "  make run            - Запустить контейнер"
	@echo "  make clean          - Очистить Docker-кэш"
	@echo "  make logs           - Посмотреть логи контейнера"
	@echo "  make shell          - Открыть shell в контейнере"
	@echo "  make bin-deps       - Установить протогенераторы (protoc-gen-go, protoc-gen-go-grpc)"
	@echo "  make protoc         - Сгенерировать gRPC-код (go-init-publisher.proto -> pkg/api/grpc)"
	@echo "  make test           - Запустить тесты"
	@echo "  make lint           - Линтинг (golangci-lint)"
	@echo "  make format         - Форматирование (go fmt ./...)"

test:
	go test ./... -v

lint:
	golangci-lint run

format:
	go fmt ./...