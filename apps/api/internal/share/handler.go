package share

import (
	"net/http"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/folder"
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
	g.POST("/shares", append(middleware, h.create)...)
	g.GET("/shares", append(middleware, h.listOutgoing)...)
	g.GET("/shares/with-me", append(middleware, h.listIncoming)...)
	g.DELETE("/shares/:id", append(middleware, h.revoke)...)
	g.GET("/shared/folders/:id", append(middleware, h.browseFolder)...)
	g.GET("/shared/files/:id/download", append(middleware, h.download)...)
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
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	Email        string `json:"email"`
}

func (h *Handler) create(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	sh, err := h.svc.Create(c.Request.Context(), userID, req.ResourceType, req.ResourceID, req.Email)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":           sh.ID,
		"resourceType": sh.ResourceType,
		"resourceId":   sh.ResourceID,
		"createdAt":    sh.CreatedAt.UTC().Format(time.RFC3339Nano),
	})
}

func (h *Handler) listOutgoing(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	out, err := h.svc.ListOutgoing(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"shares": out})
}

func (h *Handler) listIncoming(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	out, err := h.svc.ListIncoming(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"shares": out})
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

func (h *Handler) browseFolder(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	b, err := h.svc.BrowseFolder(c.Request.Context(), userID, id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, sharedBrowserJSON(b))
}

func (h *Handler) download(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.DownloadURL(c.Request.Context(), userID, id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"downloadUrl": out.URL,
		"expiresAt":   out.ExpiresAt.UTC().Format(time.RFC3339Nano),
	})
}

type sharedFolder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type sharedFile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

func sharedBrowserJSON(b folder.Browser) gin.H {
	folders := make([]sharedFolder, 0, len(b.Folders))
	for _, f := range b.Folders {
		folders = append(folders, sharedFolder{ID: f.ID, Name: f.Name})
	}
	files := make([]sharedFile, 0, len(b.Files))
	for _, f := range b.Files {
		files = append(files, sharedFile{ID: f.ID, Name: f.Name, MimeType: f.MimeType, SizeBytes: f.SizeBytes})
	}
	var current any
	if b.Folder != nil {
		current = sharedFolder{ID: b.Folder.ID, Name: b.Folder.Name}
	}
	return gin.H{
		"folder":  current,
		"folders": folders,
		"files":   files,
	}
}
