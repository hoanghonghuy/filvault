# Filnest

Personal cloud storage. Spec: [`docs/spec/README.md`](docs/spec/README.md).

## Slice 0 (hiện tại)

Cần: Docker, Go (`~/.local/go/bin/go` hoặc Go trên PATH).

```bash
cp .env.example .env
make test      # Postgres + migration test
make migrate   # áp schema lên DB trống
```

Local Postgres: `postgres://filnest:filnest@127.0.0.1:5432/filnest?sslmode=disable`
