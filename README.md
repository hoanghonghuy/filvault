# Filnest

Personal cloud storage. Spec: [`docs/spec/README.md`](docs/spec/README.md).  
Checklist bàn giao Phase 1: [`docs/spec/08-phase-1-status.md`](docs/spec/08-phase-1-status.md).

## Slice 0 (hiện tại)

Cần: Docker, Go (`~/.local/go/bin/go` hoặc Go trên PATH).

```bash
cp .env.example .env
make test      # Postgres + migration + auth tests
make migrate   # áp schema lên DB trống
make run-api   # API :8080 (invite dev-invite)
```

Local Postgres: `postgres://filnest:filnest@127.0.0.1:5432/filnest?sslmode=disable`
