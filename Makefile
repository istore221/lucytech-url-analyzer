.PHONY: compose-up compose-down compose-build compose-logs compose-ps

compose-up:
	@echo "==> Starting containers..."
	docker compose up --build -d
	@echo ""
	@echo "=========================================="
	@echo "  Frontend : http://localhost:3000"
	@echo "  Backend  : http://localhost:8080"
	@echo "=========================================="
	@echo ""

compose-down:
	@echo "==> Stopping containers..."
	docker compose down

compose-build:
	@echo "==> Building container images..."
	docker compose build

compose-logs:
	docker compose logs -f

compose-ps:
	docker compose ps