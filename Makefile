# Makefile
.PHONY: build run test clean docker-up docker-down lint setup migrate

# Go параметры
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=taskS3

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd

run:
	$(GOCMD) run ./cmd

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist/

deps:
	$(GOGET) -u ./...

lint:
	golangci-lint run

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-logs:
	docker-compose logs -f

dev: docker-up
	$(GOCMD) run ./cmd

setup:
	mkdir -p uploads/images uploads/processed/moved web/templates web/static
	@echo "✅ Структура проекта создана"

migrate:
	@echo "Создание необходимых директорий..."
	mkdir -p uploads/images uploads/processed/moved web/templates web/static

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux -v ./cmd

# Запуск с локальным MinIO
local-dev:
	@echo "Запуск локального MinIO..."
	docker run -p 9000:9000 -p 9001:9001 \
		-e "MINIO_ROOT_USER=minioadmin" \
		-e "MINIO_ROOT_PASSWORD=minioadmin" \
		-v ./minio_data:/data \
		minio/minio server /data --console-address ":9001"

# Help
help:
	@echo "Доступные команды:"
	@echo "  make build     - Собрать приложение"
	@echo "  make run       - Запустить приложение"
	@echo "  make test      - Запустить тесты"
	@echo "  make docker-up - Запустить docker-compose"
	@echo "  make dev       - Запустить в режиме разработки"
	@echo "  make setup     - Создать структуру проекта"
	@echo "  make clean     - Очистить проект"