GO ?= $(HOME)/.local/go/bin/go
COMPOSE ?= docker compose
DSN ?= postgres://filnest:filnest@127.0.0.1:5432/filnest?sslmode=disable
export PATH := $(HOME)/.local/go/bin:$(PATH)

.PHONY: compose-up compose-down migrate migrate-down test run-api

compose-up:
	$(COMPOSE) up -d postgres minio minio-init
	@echo "waiting for postgres..."
	@i=0; \
	until $(COMPOSE) exec -T postgres pg_isready -U filnest -d filnest >/dev/null 2>&1; do \
		i=$$((i+1)); \
		if [ $$i -gt 30 ]; then echo "postgres did not become ready"; exit 1; fi; \
		sleep 1; \
	done
	@echo "postgres ready"
	@echo "minio ready (bucket filnest, private)"

compose-down:
	$(COMPOSE) down

migrate: compose-up
	cd apps/api && FILNEST_DATABASE_URL="$(DSN)" $(GO) run ./cmd/migrate

migrate-down: compose-up
	cd apps/api && FILNEST_DATABASE_URL="$(DSN)" $(GO) run ./cmd/migrate -down

test: compose-up
	cd apps/api && FILNEST_DATABASE_URL="$(DSN)" $(GO) test ./... -count=1 -p 1

run-api: migrate
	cd apps/api && FILNEST_DATABASE_URL="$(DSN)" FILNEST_INVITE_CODE="dev-invite" FILNEST_JWT_SECRET="dev-jwt-secret" $(GO) run ./cmd/api
