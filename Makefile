# LinkFlow v2 - common commands
#
# Run `make` or `make help` to see available commands.

COMPOSE := docker compose -f deploy/docker-compose.yml
GO_ENV := env GOCACHE=/tmp/linkflow-go-build

.DEFAULT_GOAL := help
.PHONY: help up down restart logs ps clean reset-containers-drop-data reset-containers-keep-data psql redis-cli kafka-topics db-apply-timescale emqx-reinit redpanda-reinit redeploy-app redeploy-web hot-redeploy fullstack-test

help: ## Show this help
	@echo "LinkFlow v2 commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}'
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

reset-containers-drop-data: ## Rebuild containers and discard all data volumes
	scripts/dev/reset-containers.sh --drop-volumes

reset-containers-keep-data: ## Rebuild containers while keeping existing data volumes
	scripts/dev/reset-containers.sh --keep-volumes

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

emqx-reinit: ## Re-run emqx-init against the running stack (re-upserts auth/ACL/rules, no data loss)
	$(COMPOSE) up -d --force-recreate emqx-init

redpanda-reinit: ## Re-run redpanda-init to (idempotently) create any new Kafka topics
	$(COMPOSE) up -d --force-recreate redpanda-init

redeploy-app: ## Rebuild and restart app services only (backend, device-event-processor); keeps infra and data
	$(COMPOSE) up -d --build --force-recreate backend device-event-processor

redeploy-web: ## Rebuild and restart the web container (multi-stage pnpm build → nginx)
	$(COMPOSE) up -d --build --force-recreate web

hot-redeploy: db-apply-timescale redeploy-app redeploy-web emqx-reinit redpanda-reinit ## Apply SQL + rebuild app + rebuild web + re-init EMQX + create new Kafka topics, without dropping data

fullstack-test: ## Run web dev server against the compose backend/services (uses .nvmrc via nvm)
	@echo "Starting web dev server. Backend services run in docker compose. Press Ctrl-C to stop."
	@bash -c '. "$$HOME/.nvm/nvm.sh" && cd web && nvm use && pnpm dev'


.PHONY: go fmt fmt-list fmt-fix vet test ci
GO_FILES := $(shell find . -name '*.go' -not -path '*/vendor/*')
GO_MODULES := pkg service/mqtt-gateway service/device-event-processor service/backend

fmt: ## Check Go formatting
	@test -z "$$(gofmt -l $(GO_FILES))"

fmt-list: ## List unformatted Go files
	@gofmt -l $(GO_FILES)

fmt-fix: ## Format Go files
	@gofmt -w $(GO_FILES)

vet: ## Run Go vet
	@for m in $(GO_MODULES); do \
		echo "==> go vet $$m"; \
		(cd $$m && $(GO_ENV) go vet ./...); \
	done

test: ## Run Go tests
	@for m in $(GO_MODULES); do \
		echo "==> go test $$m"; \
		(cd $$m && $(GO_ENV) go test ./...); \
	done

ci: fmt vet test ## Run local CI checks
