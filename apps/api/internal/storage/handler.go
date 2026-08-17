package storage

import (
	"context"
	"net/http"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type QuotaReader interface {
	GetStorage(ctx context.Context, userID string) (used, quota int64, err error)
}

type Handler struct {
	quota QuotaReader
}

func NewHandler(quota QuotaReader) *Handler {
	return &Handler{quota: quota}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.GET("/storage", append(middleware, h.get)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

func (h *Handler) get(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	used, quota, err := h.quota.GetStorage(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"usedBytes":  used,
		"quotaBytes": quota,
	})
}
