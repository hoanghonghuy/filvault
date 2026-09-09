package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"filvault/internal/app"
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type readinessBody struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func TestHealthz_ReturnsOKWithoutDatabaseDependency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := app.NewWithDeps(config.Config{}, nil, mailer.NewMemory(), objectstore.NewMemory())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v body=%s", err, rec.Body.String())
	}
	if body.Status != "ok" {
		t.Fatalf("status field=%q", body.Status)
	}
}

func TestMetrics_ReturnsPrometheusCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	pool := openPool(t, ctx)
	t.Cleanup(pool.Close)
	engine := app.NewWithDeps(config.Config{}, pool, mailer.NewMemory(), objectstore.NewMemory())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "filvault_http_requests_total") {
		t.Fatalf("metrics status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReadyz_ReturnsReadyWhenDependenciesAreHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	pool := openPool(t, ctx)
	t.Cleanup(pool.Close)
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	engine := app.NewWithDeps(config.Config{}, pool, mailer.NewMemory(), objectstore.NewMemory())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("readyz status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReadiness(t, rec.Body.Bytes())
	if body.Status != "ready" {
		t.Fatalf("status=%q", body.Status)
	}
	assertCheck(t, body.Checks, "database", "ok")
	assertCheck(t, body.Checks, "schema", "ok")
	assertCheck(t, body.Checks, "object_store", "ok")
}

func TestReadyz_ReturnsNotReadyWhenDatabaseIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	pool := openPool(t, ctx)
	t.Cleanup(pool.Close)
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool.Close()

	engine := app.NewWithDeps(config.Config{}, pool, mailer.NewMemory(), objectstore.NewMemory())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReadiness(t, rec.Body.Bytes())
	if body.Status != "not_ready" {
		t.Fatalf("status=%q", body.Status)
	}
	assertCheck(t, body.Checks, "database", "failed")
}

func TestReadyz_ReturnsNotReadyWhenSchemaIsIncomplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	pool := openPool(t, ctx)
	t.Cleanup(pool.Close)

	engine := app.NewWithDeps(config.Config{}, pool, mailer.NewMemory(), objectstore.NewMemory())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReadiness(t, rec.Body.Bytes())
	if body.Status != "not_ready" {
		t.Fatalf("status=%q", body.Status)
	}
	assertCheck(t, body.Checks, "schema", "failed")
}

func TestReadyz_ReturnsNotReadyWhenObjectStoreIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	pool := openPool(t, ctx)
	t.Cleanup(pool.Close)
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	engine := app.NewWithDeps(
		config.Config{},
		pool,
		mailer.NewMemory(),
		objectstore.NewUnavailable(nil),
	)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReadiness(t, rec.Body.Bytes())
	if body.Status != "not_ready" {
		t.Fatalf("status=%q", body.Status)
	}
	assertCheck(t, body.Checks, "object_store", "failed")
}

func decodeReadiness(t *testing.T, raw []byte) readinessBody {
	t.Helper()
	var body readinessBody
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json: %v body=%s", err, string(raw))
	}
	return body
}

func assertCheck(t *testing.T, checks map[string]string, name, want string) {
	t.Helper()
	got, ok := checks[name]
	if !ok {
		t.Fatalf("missing check %q in %#v", name, checks)
	}
	if got != want {
		t.Fatalf("check %q=%q want %q", name, got, want)
	}
}

func openPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("FILVAULT_DATABASE_URL")
	if url == "" {
		t.Fatal("FILVAULT_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return pool
}
