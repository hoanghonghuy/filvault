package file

import (
	"encoding/json"
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

func (h *Handler) RegisterRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.POST("/files/upload-sessions", append(middleware, h.createSession)...)
	g.POST("/files/:id/complete", append(middleware, h.complete)...)
	g.GET("/files/:id", append(middleware, h.get)...)
	g.GET("/files/:id/download", append(middleware, h.download)...)
	g.PATCH("/files/:id", append(middleware, h.patch)...)
	g.DELETE("/files/:id", append(middleware, h.delete)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

type sessionReq struct {
	Name        string  `json:"name"`
	Size        int64   `json:"size"`
	ContentType string  `json:"contentType"`
	FolderID    *string `json:"folderId"`
}

func (h *Handler) createSession(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req sessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	session, err := h.svc.CreateUploadSession(c.Request.Context(), userID, req.Name, req.ContentType, req.Size, req.FolderID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"fileId":    session.FileID,
		"uploadUrl": session.UploadURL,
		"expiresAt": session.ExpiresAt.UTC().Format(time.RFC3339Nano),
	})
}

func (h *Handler) complete(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	f, err := h.svc.Complete(c.Request.Context(), userID, id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicFrom(f))
}

func (h *Handler) get(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	f, err := h.svc.Get(c.Request.Context(), userID, id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicFrom(f))
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

func (h *Handler) delete(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) patch(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		httpx.Validation(c)
		return
	}
	var name *string
	if v, ok := raw["name"]; ok {
		var n string
		if err := json.Unmarshal(v, &n); err != nil {
			httpx.Validation(c)
			return
		}
		name = &n
	}
	moveFolder := false
	var newFolder *string
	if v, ok := raw["folderId"]; ok {
		moveFolder = true
		if string(v) == "null" {
			newFolder = nil
		} else {
			var id string
			if err := json.Unmarshal(v, &id); err != nil {
				httpx.Validation(c)
				return
			}
			newFolder = &id
		}
	}
	if name == nil && !moveFolder {
		httpx.Validation(c)
		return
	}
	id, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	f, err := h.svc.Patch(c.Request.Context(), userID, id, name, moveFolder, newFolder)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicFrom(f))
}

type publicFile struct {
	ID        string  `json:"id"`
	FolderID  *string `json:"folderId"`
	Name      string  `json:"name"`
	MimeType  string  `json:"mimeType"`
	SizeBytes int64   `json:"sizeBytes"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

func publicFrom(f File) publicFile {
	return publicFile{
		ID:        f.ID,
		FolderID:  f.FolderID,
		Name:      f.Name,
		MimeType:  f.MimeType,
		SizeBytes: f.SizeBytes,
		Status:    f.Status,
		CreatedAt: f.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: f.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
