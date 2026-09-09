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

func readiness(pool *pgxpool.Pool, store objectstore.ObjectStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessCheckTimeout)
		defer cancel()

		checks := map[string]string{}
		ready := true

		if err := pool.Ping(ctx); err != nil {
			checks["database"] = "failed"
			ready = false
		} else {
			checks["database"] = "ok"
		}

		if err := postgres.SchemaReady(ctx, pool); err != nil {
			checks["schema"] = "failed"
			ready = false
		} else {
			checks["schema"] = "ok"
		}

		if err := store.CheckReady(ctx); err != nil {
			checks["object_store"] = "failed"
			ready = false
		} else {
			checks["object_store"] = "ok"
		}

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
