package sharelink

import (
	"net/http"
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

// RegisterOwnerRoutes mounts the authenticated owner endpoints.
func (h *Handler) RegisterOwnerRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.POST("/files/:id/share", append(middleware, h.create)...)
	g.DELETE("/files/:id/share", append(middleware, h.revoke)...)
	g.GET("/share-links", append(middleware, h.list)...)
}

// RegisterPublicRoutes mounts the anonymous endpoints behind the given
// middlewares (rate limiting is wired by the app composition root).
func (h *Handler) RegisterPublicRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.GET("/public/shares/:token", append(middleware, h.publicMeta)...)
	g.GET("/public/shares/:token/download", append(middleware, h.publicDownload)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

type createReq struct {
	ExpiresIn string `json:"expiresIn"`
}

func (h *Handler) create(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	var req createReq
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Validation(c)
			return
		}
	}
	out, err := h.svc.Create(c.Request.Context(), userID, id, req.ExpiresIn)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":        out.ID,
		"fileId":    out.FileID,
		"url":       out.URL,
		"token":     out.Token,
		"expiresAt": formatTimePtr(out.ExpiresAt),
		"createdAt": out.CreatedAt.UTC().Format(time.RFC3339Nano),
	})
}

func (h *Handler) revoke(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Revoke(c.Request.Context(), userID, id); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) list(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	links, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(links))
	for _, l := range links {
		out = append(out, gin.H{
			"id":        l.ID,
			"fileId":    l.FileID,
			"fileName":  l.FileInfo.Name,
			"mimeType":  l.FileInfo.MimeType,
			"sizeBytes": l.FileInfo.SizeBytes,
			"expiresAt": formatTimePtr(l.ExpiresAt),
			"createdAt": l.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	c.JSON(http.StatusOK, gin.H{"links": out})
}

func (h *Handler) publicMeta(c *gin.Context) {
	meta, err := h.svc.PublicMeta(c.Request.Context(), c.Param("token"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name":      meta.Name,
		"mimeType":  meta.MimeType,
		"sizeBytes": meta.SizeBytes,
		"expiresAt": formatTimePtr(meta.ExpiresAt),
	})
}

func (h *Handler) publicDownload(c *gin.Context) {
	out, err := h.svc.PublicDownload(c.Request.Context(), c.Param("token"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"downloadUrl": out.URL,
		"expiresAt":   out.ExpiresAt.UTC().Format(time.RFC3339Nano),
	})
}

func formatTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}
