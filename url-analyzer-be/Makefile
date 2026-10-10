IMAGE_NAME := lucytech-url-analyzer
DOCKERFILE := Dockerfile

.PHONY: all help build run clean


all: run

quality:
	@echo "==> Running tests and generating coverage report..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage Report per File:"
	@go tool cover -func=coverage.out

run: build
	@echo "==> Running application in Docker..."
	@docker run --rm $(IMAGE_NAME)


build:
	@echo "==> Building Docker image..."
	@docker build -t $(IMAGE_NAME) -f $(DOCKERFILE) .


clean:
	@echo "==> Cleaning up Docker image..."
	@docker rmi $(IMAGE_NAME) || true