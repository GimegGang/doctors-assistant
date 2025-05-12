APP_NAME ?= kode-app
DOCKER_IMAGE ?= kode-image
DOCKERFILE ?= main.dockerfile
GO_FILES ?= $(shell find . -type f -name '*.go' -not -path "./vendor/*")
UNIT_TEST_PKGS ?= $(shell go list ./... | grep -v /tests/)
INTEGRATION_TEST_PKGS ?= ./tests/...


# Сборка приложения
build:
	@echo "Building application..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/kode/main.go

# Сборка Docker-образа
docker-build:
	@echo "Building Docker image from $(DOCKERFILE)..."
	@docker build -t $(DOCKER_IMAGE) -f $(DOCKERFILE) .

# Запуск в контейнере
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -p 1234:1234 --name $(APP_NAME) $(DOCKER_IMAGE)

# Запуск юнит-тестов
unit-test:
	@echo "Running unit tests..."
	@go test -v -short $(UNIT_TEST_PKGS)

# Запуск интеграционных тестов
integration-test:
	@echo "Running integration tests..."
	@go test -v $(INTEGRATION_TEST_PKGS)

# Запуск линтера
lint:
	@echo "Running linter..."
	@golangci-lint run --config .golangci.yml

# Запуск приложения локально
run:
	@echo "Starting application..."
	@go run ./cmd/kode/main.go

# Очистка
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@docker rm -f $(APP_NAME) || true
	@docker rmi $(DOCKER_IMAGE) || true
