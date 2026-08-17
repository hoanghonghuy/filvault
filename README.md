# Filnest

Personal cloud storage. Spec: [`docs/spec/README.md`](docs/spec/README.md).  
Checklist bàn giao Phase 1: [`docs/spec/08-phase-1-status.md`](docs/spec/08-phase-1-status.md).

## Chạy local

```bash
cp .env.example .env
make test          # Postgres + MinIO + migration + API tests
make migrate       # áp schema (chỉ infra)
make up            # Docker: postgres + minio + api + web
make down          # dừng toàn bộ stack
make dev-api       # API trên host :8080
make dev-web       # Vue trên host :5173 (cần Node ^22)
```

Sau `make up`: web `http://localhost:5173`, API `http://localhost:8080`, MinIO `http://localhost:9000`.
