# LinkFlow v2 - common commands
#
# Run `make` or `make help` to see available commands.

COMPOSE := docker compose -f deploy/docker-compose.yml

.DEFAULT_GOAL := help
.PHONY: help up down restart logs ps clean psql redis-cli kafka-topics

help: ## Show this help
	@echo "LinkFlow v2 commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'
	@echo ""

# --- Lifecycle ---

up: ## Start all infrastructure services
	$(COMPOSE) up -d

down: ## Stop all infrastructure services (keeps data)
	$(COMPOSE) down

restart: ## Restart all services
	$(COMPOSE) restart

clean: ## Stop and remove all data (DESTRUCTIVE - wipes volumes)
	$(COMPOSE) down -v

# --- Inspection ---

ps: ## Show running services and health status
	$(COMPOSE) ps

logs: ## Tail logs from all services
	$(COMPOSE) logs -f

# --- Service shortcuts ---

psql: ## Open psql shell on TimescaleDB
	docker exec -it linkflow-timescaledb psql -U linkflow -d linkflow

redis-cli: ## Open redis-cli on Redis
	docker exec -it linkflow-redis redis-cli

kafka-topics: ## List Kafka topics
	docker exec -it linkflow-redpanda rpk topic list


.PHONY: mqtt-service fmt fmt-fix vet test ci

fmt: ## Check Go formatting
	cd service/mqtt-gateway && test -z "$$(gofmt -l .)"

fmt-list: ## List unformatted Go files
	cd service/mqtt-gateway && gofmt -l .

fmt-fix: ## Format Go files
	cd service/mqtt-gateway && gofmt -w .

vet: ## Run Go vet
	cd service/mqtt-gateway && go vet ./...

test: ## Run Go tests
	cd service/mqtt-gateway && go test ./...

ci: fmt vet test ## Run local CI checks




