# LinkFlow v2 - common commands
#
# Run `make` or `make help` to see available commands.

COMPOSE := docker compose -f deploy/docker-compose.yml

.DEFAULT_GOAL := help
.PHONY: help up down restart logs ps clean psql redis-cli kafka-topics db-apply-timescale service-test

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

db-apply-timescale: ## Apply TimescaleDB SQL files to the existing database
	scripts/db/apply-timescaledb.sh

service-test: ## Run mqtt-gateway and device-event-processor from the repo root
	@echo "Starting device-event-processor and mqtt-gateway. Press Ctrl-C to stop both."
	@set -e; \
	( cd service/device-event-processor && env GOCACHE=/tmp/linkflow-go-build go run ./cmd ) & \
	processor_pid=$$!; \
	( cd service/mqtt-gateway && env GOCACHE=/tmp/linkflow-go-build MQTT_CLIENT_ID=linkflow-mqtt-gateway-service-test MQTT_CLEAN_SESSION=true go run ./cmd ) & \
	gateway_pid=$$!; \
	trap 'kill $$processor_pid $$gateway_pid 2>/dev/null || true' INT TERM EXIT; \
	wait $$processor_pid $$gateway_pid


.PHONY: go fmt fmt-list fmt-fix vet test ci
GO_FILES := $(shell find . -name '*.go' -not -path '*/vendor/*')
GO_MODULES := pkg service/mqtt-gateway service/device-event-processor

fmt: ## Check Go formatting
	@test -z "$$(gofmt -l $(GO_FILES))"

fmt-list: ## List unformatted Go files
	@gofmt -l $(GO_FILES)

fmt-fix: ## Format Go files
	@gofmt -w $(GO_FILES)

vet: ## Run Go vet
	@for m in $(GO_MODULES); do \
		echo "==> go vet $$m"; \
		(cd $$m && go vet ./...); \
	done

test: ## Run Go tests
	@for m in $(GO_MODULES); do \
		echo "==> go test $$m"; \
		(cd $$m && go test ./...); \
	done

ci: fmt vet test ## Run local CI checks
