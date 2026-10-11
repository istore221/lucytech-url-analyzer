IMAGE_NAME := lucytech-url-analyzer
DOCKERFILE := Dockerfile
# Default URL if none is provided on the command line
URL := https://example.com

.PHONY: all help build run clean


all: run

quality:
	@echo "==> Running linter..."
	@golangci-lint run ./...
	@echo "==> Running tests and generating raw coverage report..."
	@go test -v -coverprofile=coverage.raw.out ./...

	@echo "==> Filtering out main.go and mocks..."
	@grep -v -E "main\.go|/mocks/" coverage.raw.out > coverage.out

	@echo "==> Generating HTML report..."
	@go tool cover -html=coverage.out -o coverage.html

	@echo "==> Coverage Report per File:"
	@go tool cover -func=coverage.out

	@rm -f coverage.raw.out

run: build
	@echo "==> Running application in Docker..."
	@docker run --rm $(IMAGE_NAME) "$(URL)"


build:
	@echo "==> Building Docker image..."
	@docker build -t $(IMAGE_NAME) -f $(DOCKERFILE) .


clean:
	@echo "==> Cleaning up Docker image..."
	@docker rmi $(IMAGE_NAME) || true