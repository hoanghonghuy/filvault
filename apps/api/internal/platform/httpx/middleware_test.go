package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestCORS_PreflightAllowedOrigin(t *testing.T) {
	engine := gin.New()
	engine.Use(httpx.CORS([]string{"http://localhost:5173"}))
	engine.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "authorization, last-event-id")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow-origin=%q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type, Last-Event-ID" {
		t.Fatalf("allow-headers=%q", got)
	}
}

func TestCORS_UnknownOrigin(t *testing.T) {
	engine := gin.New()
	engine.Use(httpx.CORS([]string{"http://localhost:5173"}))
	engine.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected CORS header for unknown origin")
	}
}

func TestCORS_PrivateLANOrigin(t *testing.T) {
	engine := gin.New()
	engine.Use(httpx.CORS([]string{"http://localhost:5173"}))
	engine.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://192.168.1.50:5173")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://192.168.1.50:5173" {
		t.Fatalf("allow-origin=%q", got)
	}
}

func TestCORS_TryCloudflareOrigin(t *testing.T) {
	engine := gin.New()
	engine.Use(httpx.CORS([]string{"http://localhost:5173"}))
	engine.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://quick-tunnel-subdomain.trycloudflare.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://quick-tunnel-subdomain.trycloudflare.com" {
		t.Fatalf("allow-origin=%q", got)
	}
}
