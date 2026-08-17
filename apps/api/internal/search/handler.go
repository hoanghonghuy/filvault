package search

import (
	"net/http"
	"strconv"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.GET("/search", append(middleware, h.search)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

func (h *Handler) search(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	q := c.Query("q")
	limit := DefaultLimit
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			httpx.Validation(c)
			return
		}
		limit = n
	}
	out, err := h.svc.Search(c.Request.Context(), userID, q, limit)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicResult(out))
}

func publicResult(r Result) gin.H {
	folders := make([]gin.H, 0, len(r.Folders))
	for _, f := range r.Folders {
		folders = append(folders, gin.H{
			"id":        f.ID,
			"parentId":  f.ParentID,
			"name":      f.Name,
			"createdAt": f.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updatedAt": f.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	files := make([]gin.H, 0, len(r.Files))
	for _, f := range r.Files {
		files = append(files, gin.H{
			"id":        f.ID,
			"name":      f.Name,
			"mimeType":  f.MimeType,
			"sizeBytes": f.SizeBytes,
			"updatedAt": f.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return gin.H{"folders": folders, "files": files}
}
