GO ?= $(HOME)/.local/go/bin/go
COMPOSE ?= docker compose
PROD_COMPOSE = $(COMPOSE) -f docker-compose.prod.yml --env-file deploy/production/.env

ifneq (,$(wildcard .env))
include .env
export
endif

FILVAULT_POSTGRES_PASSWORD ?= replace-with-local-postgres-password
DSN ?= postgres://filvault:$(FILVAULT_POSTGRES_PASSWORD)@127.0.0.1:5435/filvault?sslmode=disable
export PATH := $(HOME)/.local/go/bin:$(PATH)
export FILVAULT_POSTGRES_PASSWORD

.PHONY: compose-up compose-down up down prod-up prod-like-up prod-like-down migrate migrate-down seed test run-api dev-api dev-web \
	lint-api typecheck-api lint-web typecheck-web ci-api ci-web cli-build cli-test secret-scan e2e

compose-up:
	$(COMPOSE) up -d postgres minio minio-init
	@echo "waiting for postgres..."
	@i=0; \
	until $(COMPOSE) exec -T postgres pg_isready -U filvault -d filvault >/dev/null 2>&1; do \
		i=$$((i+1)); \
		if [ $$i -gt 30 ]; then echo "postgres did not become ready"; exit 1; fi; \
		sleep 1; \
	done
	@echo "postgres ready"
	@echo "minio ready (bucket filvault, private)"

compose-down:
	$(COMPOSE) down

up:
	$(COMPOSE) up -d --build

prod-up:
	$(COMPOSE) up -d --build migrate api worker web

prod-like-up:
	@test -f deploy/production/.env || (echo "run ./deploy/production/generate-env.sh first" >&2; exit 1)
	./deploy/production/postgres/generate-certs.sh
	$(PROD_COMPOSE) up -d --build
	@echo "waiting for API /readyz..."
	@i=0; \
	until $(PROD_COMPOSE) exec -T api wget -qO- http://127.0.0.1:8080/readyz 2>/dev/null | grep -q '"status":"ready"'; do \
		i=$$((i+1)); \
		if [ $$i -gt 60 ]; then echo "API did not become ready"; $(PROD_COMPOSE) logs --tail=40 api; exit 1; fi; \
		sleep 2; \
	done
	@echo "production-like stack ready at https://$${FILVAULT_PUBLIC_HOST:-filvault.local}:$${FILVAULT_PUBLIC_PORT:-8443}"

prod-like-down:
	$(PROD_COMPOSE) down

down:
	$(COMPOSE) down

migrate: compose-up
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/migrate

migrate-down: compose-up
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/migrate -down

seed: migrate
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/seed

test: compose-up
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) test ./... -count=1 -p 1

run-api: seed
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/api

dev-api: seed
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/api

dev-web:
	cd apps/web && npm run dev

lint-api:
	cd apps/api && $(GO) vet ./...

typecheck-api:
	cd apps/api && $(GO) build -o /dev/null ./...

lint-web:
	cd apps/web && npm run lint

typecheck-web:
	cd apps/web && npm run type-check

ci-api: lint-api typecheck-api test

ci-web: lint-web typecheck-web

cli-build:
	cd apps/cli && $(GO) build -o ../../bin/filvault ./cmd/filvault

cli-test:
	cd apps/cli && $(GO) test ./... -count=1

secret-scan:
	./scripts/verify-secret-scan-gate.sh

e2e:
	./scripts/run-e2e.sh
