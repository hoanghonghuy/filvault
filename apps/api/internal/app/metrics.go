package app

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type requestMetrics struct {
	requests atomic.Uint64
	errors   atomic.Uint64
	active   atomic.Int64
}

func newRequestMetrics() *requestMetrics {
	return &requestMetrics{}
}

func (m *requestMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.active.Add(1)
		defer m.active.Add(-1)
		m.requests.Add(1)
		c.Next()
		if c.Writer.Status() >= http.StatusInternalServerError {
			m.errors.Add(1)
		}
	}
}

func (m *requestMetrics) Handler(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(c.Writer,
		"# HELP filvault_http_requests_total Total HTTP requests.\n"+
			"# TYPE filvault_http_requests_total counter\n"+
			"filvault_http_requests_total %d\n"+
			"# HELP filvault_http_errors_total Total HTTP 5xx responses.\n"+
			"# TYPE filvault_http_errors_total counter\n"+
			"filvault_http_errors_total %d\n"+
			"# HELP filvault_http_active_requests Current active HTTP requests.\n"+
			"# TYPE filvault_http_active_requests gauge\n"+
			"filvault_http_active_requests %d\n"+
			"# HELP filvault_metrics_scrape_timestamp_seconds Unix time of this scrape.\n"+
			"# TYPE filvault_metrics_scrape_timestamp_seconds gauge\n"+
			"filvault_metrics_scrape_timestamp_seconds %d\n",
		m.requests.Load(), m.errors.Load(), m.active.Load(), time.Now().Unix(),
	)
}
