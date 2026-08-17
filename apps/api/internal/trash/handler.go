package trash

import (
	"net/http"
	"strings"

	"filnest/internal/apperr"
	"filnest/internal/auth"
	"filnest/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.GET("/trash", append(middleware, h.list)...)
	g.DELETE("/trash/:type/:id", append(middleware, h.permanentDelete)...)
	g.POST("/files/:id/restore", append(middleware, h.restoreFile)...)
	g.POST("/folders/:id/restore", append(middleware, h.restoreFolder)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

func (h *Handler) list(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	out, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) restoreFile(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	if err := h.svc.RestoreFile(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) restoreFolder(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	if err := h.svc.RestoreFolder(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) permanentDelete(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	typ := strings.ToLower(c.Param("type"))
	id := c.Param("id")
	var err error
	switch typ {
	case "files":
		err = h.svc.PermanentDeleteFile(c.Request.Context(), userID, id)
	case "folders":
		err = h.svc.PermanentDeleteFolder(c.Request.Context(), userID, id)
	default:
		httpx.Validation(c)
		return
	}
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
