package app

import (
	"context"
	"net/http"
	"time"

	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const readinessCheckTimeout = 2 * time.Second

func healthz() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

type dependencyCheck struct {
	name string
	run  func(ctx context.Context) error
}

func readiness(pool *pgxpool.Pool, store objectstore.ObjectStore) gin.HandlerFunc {
	checker := readinessChecker{
		checks: []dependencyCheck{
			{name: "database", run: pool.Ping},
			{name: "schema", run: func(ctx context.Context) error {
				return postgres.SchemaReady(ctx, pool)
			}},
			{name: "object_store", run: store.CheckReady},
		},
	}
	return func(c *gin.Context) {
		ready, checks := checker.evaluate(c.Request.Context())
		status := "ready"
		code := http.StatusOK
		if !ready {
			status = "not_ready"
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{
			"status": status,
			"checks": checks,
		})
	}
}

type readinessChecker struct {
	checks []dependencyCheck
}

func (r readinessChecker) evaluate(parent context.Context) (bool, map[string]string) {
	checks := make(map[string]string, len(r.checks))
	ready := true
	for _, check := range r.checks {
		ctx, cancel := context.WithTimeout(parent, readinessCheckTimeout)
		err := check.run(ctx)
		cancel()
		if err != nil {
			checks[check.name] = "failed"
			ready = false
			continue
		}
		checks[check.name] = "ok"
	}
	return ready, checks
}
