package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"filvault/internal/apperr"

	"github.com/gin-gonic/gin"
)

func Error(c *gin.Context, err error) {
	var e *apperr.Error
	if errors.As(err, &e) {
		c.JSON(e.HTTPStatus, gin.H{
			"error": gin.H{
				"code":    e.Code,
				"message": e.Message,
			},
		})
		return
	}
	slog.Error("request failed", "err", err, "path", c.Request.URL.Path)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "INTERNAL",
			"message": "Internal error",
		},
	})
}

func Validation(c *gin.Context) {
	Error(c, apperr.Validation)
}
