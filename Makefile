.PHONY: run stop build logs ps

run:
	@echo "==> Starting containers..."
	docker compose up --build -d
	@echo ""
	@echo "=========================================="
	@echo "  Frontend : http://localhost:3000"
	@echo "  Backend  : http://localhost:8080"
	@echo "=========================================="
	@echo ""

stop:
	@echo "==> Stopping containers..."
	docker compose down

build:
	@echo "==> Building container images..."
	docker compose build

logs:
	docker compose logs -f

ps:
	docker compose ps