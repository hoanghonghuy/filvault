GO ?= $(HOME)/.local/go/bin/go
COMPOSE ?= docker compose
DSN ?= postgres://filvault:filvault@127.0.0.1:5432/filvault?sslmode=disable
export PATH := $(HOME)/.local/go/bin:$(PATH)

.PHONY: compose-up compose-down up down migrate migrate-down seed test run-api dev-api dev-web \
	lint-api typecheck-api lint-web typecheck-web ci-api ci-web

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

down:
	$(COMPOSE) down

migrate: compose-up
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/migrate

migrate-down: compose-up
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) run ./cmd/migrate -down

seed: migrate
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" FILVAULT_INVITE_CODE="dev-invite" $(GO) run ./cmd/seed

test: compose-up
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" $(GO) test ./... -count=1 -p 1

run-api: seed
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" FILVAULT_INVITE_CODE="dev-invite" FILVAULT_JWT_SECRET="dev-jwt-secret" FILVAULT_S3_ENDPOINT="http://127.0.0.1:9000" FILVAULT_S3_ACCESS_KEY="filvault" FILVAULT_S3_SECRET_KEY="filvaultsecret" $(GO) run ./cmd/api

dev-api: seed
	cd apps/api && FILVAULT_DATABASE_URL="$(DSN)" FILVAULT_INVITE_CODE="dev-invite" FILVAULT_JWT_SECRET="dev-jwt-secret" FILVAULT_S3_ENDPOINT="http://127.0.0.1:9000" FILVAULT_S3_ACCESS_KEY="filvault" FILVAULT_S3_SECRET_KEY="filvaultsecret" $(GO) run ./cmd/api

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
