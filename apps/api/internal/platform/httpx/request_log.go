package httpx

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			slog.String("clientIp", c.ClientIP()),
		}
		if v, ok := c.Get("authUserID"); ok {
			if id, _ := v.(string); id != "" {
				attrs = append(attrs, slog.String("userId", id))
			}
		}
		if err := c.Errors.Last(); err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		level := slog.LevelInfo
		status := c.Writer.Status()
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(c.Request.Context(), level, "http request", attrs...)
	}
}
