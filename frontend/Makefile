# Makefile для фронтенд-проекта на Windows

# Указываем командную оболочку на Windows
SHELL := cmd.exe

# Имя Docker образа
IMAGE_NAME := frontend-template-generator
CONTAINER_NAME := frontend-template-generator-container

# Порт для разработки
PORT := 5173

.PHONY: help setup dev build clean docker-setup docker-dev docker-build docker-clean windows-setup windows-dev windows-build windows-clean windows-rebuild windows-clean-deps fix-deps cors-proxy dev-cors

# Вывод доступных команд
help:
	@echo Доступные команды:
	@echo - setup: Настройка проекта локально (требует npm)
	@echo - dev: Локальный запуск сервера разработки (требует npm)
	@echo - build: Локальная сборка проекта (требует npm)
	@echo - clean: Очистка локальных артефактов сборки
	@echo - docker-setup: Сборка Docker образа
	@echo - docker-dev: Запуск сервера разработки в Docker
	@echo - docker-build: Сборка проекта в Docker
	@echo - docker-clean: Остановка и удаление контейнера
	@echo - windows-setup: Настройка проекта через PowerShell
	@echo - windows-dev: Запуск сервера разработки через PowerShell
	@echo - windows-build: Сборка проекта через PowerShell
	@echo - windows-clean: Очистка через PowerShell
	@echo - windows-rebuild: Перестроить и перезапустить контейнер через PowerShell
	@echo - windows-clean-deps: Очистить кэш зависимостей Vite через PowerShell
	@echo - fix-deps: Исправить проблемы с зависимостями
	@echo - cors-proxy: Запустить CORS прокси сервер
	@echo - dev-cors: Запустить разработку с CORS прокси

# Переменные для работы с Docker
DOCKER_COMPOSE = docker-compose -f docker/docker-compose.yml

# Локальные команды (требуют установленного npm)
setup:
	npm ci

dev:
	npm run dev

build:
	npm run build

clean:
	rm -rf dist
	rm -rf node_modules

# Команда для запуска CORS прокси
cors-proxy:
	node cors-proxy.js

# Команда для запуска разработки с CORS прокси
dev-cors:
	npm run dev:cors

# Команды для исправления зависимостей
fix-deps:
	npm run clean
	npm run reset
	npm run dev:clear

# Docker команды
docker-setup:
	@echo Сборка Docker образа...
	docker build -t $(IMAGE_NAME) -f docker/Dockerfile .

docker-dev:
	@echo Запуск контейнера...
	docker run -d --name $(CONTAINER_NAME) -p $(PORT):$(PORT) -v "%cd%":/app --add-host=host.docker.internal:host-gateway $(IMAGE_NAME) npm run dev -- --host 0.0.0.0
	@echo Сервер доступен по адресу: http://localhost:$(PORT)
	@echo Для остановки используйте: make docker-clean

docker-build: docker-setup
	@echo Сборка проекта в Docker...
	docker run --rm -v "%cd%":/app $(IMAGE_NAME) npm run build

docker-clean:
	@echo Остановка контейнера...
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true

# Команды Docker Compose
compose-setup:
	$(DOCKER_COMPOSE) build

compose-dev:
	$(DOCKER_COMPOSE) up

compose-build:
	$(DOCKER_COMPOSE) run --rm frontend npm run build

compose-clean:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) rm -f

# Windows-совместимые команды (через PowerShell)
windows-setup:
	powershell -Command "docker-compose -f docker\docker-compose.yml build"

windows-dev:
	powershell -Command "docker-compose -f docker\docker-compose.yml up"

windows-build:
	powershell -Command "docker-compose -f docker\docker-compose.yml run --rm frontend npm run build"

windows-clean:
	powershell -Command "docker-compose -f docker\docker-compose.yml down"
	powershell -Command "docker-compose -f docker\docker-compose.yml rm -f"

# Команда для перестройки и перезапуска контейнера
windows-rebuild:
	powershell -Command "docker-compose -f docker\docker-compose.yml down"
	powershell -Command "docker-compose -f docker\docker-compose.yml build --no-cache"
	powershell -Command "docker-compose -f docker\docker-compose.yml up"

# Очистка кэша зависимостей Vite в Docker
windows-clean-deps:
	powershell -Command "docker-compose -f docker\docker-compose.yml exec frontend npm run clean"
	powershell -Command "docker-compose -f docker\docker-compose.yml exec frontend npm run reset" 