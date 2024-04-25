.PHONY: build run lint docker-build

DOCKER_COMPOSE = docker-compose
GOLINT = golangci-lint

build:
	@echo "Building the project..."
	go build -v ./...

docker-build:
	@echo "Building and running with Docker Compose..."
	$(DOCKER_COMPOSE) up --build

run:
	@echo "Running with Docker Compose..."
	$(DOCKER_COMPOSE) up

down:
	@echo "Stopping and removing containers..."
	$(DOCKER_COMPOSE) down

lint:
	@echo "Linting the Go files..."
	$(GOLINT) run ./...

vet:
	@echo "Running go vet..."
	go vet ./...

clean:
	@echo "Cleaning up..."
	go clean
	rm -f myapp

help:
	@echo "Makefile commands:"
	@echo "build        - Build the Go application."
	@echo "docker-build - Build and run the containers using Docker Compose."
	@echo "run          - Run the existing containers."
	@echo "down         - Stop and remove containers."
	@echo "lint         - Lint the Go source code."
	@echo "vet          - Run go vet."
	@echo "clean        - Clean the build artifacts."