package folder

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	g.POST("/folders", append(middleware, h.create)...)
	g.GET("/folders/:id", append(middleware, h.get)...)
	g.PATCH("/folders/:id", append(middleware, h.patch)...)
	g.DELETE("/folders/:id", append(middleware, h.delete)...)
	g.GET("/browser", append(middleware, h.browser)...)
}

const ctxUserID = auth.CtxUserID

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

type createReq struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parentId"`
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
	f, err := h.svc.Create(c.Request.Context(), userID, req.Name, req.ParentID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, publicFrom(f))
}

func (h *Handler) get(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	f, err := h.svc.Get(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicFrom(f))
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
	moveParent := false
	var newParent *string
	if v, ok := raw["parentId"]; ok {
		moveParent = true
		if string(v) == "null" {
			newParent = nil
		} else {
			var p string
			if err := json.Unmarshal(v, &p); err != nil {
				httpx.Validation(c)
				return
			}
			newParent = &p
		}
	}
	if name == nil && !moveParent {
		httpx.Validation(c)
		return
	}
	f, err := h.svc.Patch(c.Request.Context(), userID, c.Param("id"), name, moveParent, newParent)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicFrom(f))
}

func (h *Handler) delete(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) browser(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var folderID *string
	if q := strings.TrimSpace(c.Query("folderId")); q != "" {
		folderID = &q
	}
	result, err := h.svc.Browser(c.Request.Context(), userID, folderID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, browserJSON(result))
}

type publicFolder struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parentId"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type publicFile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
	UpdatedAt string `json:"updatedAt"`
}

func publicFrom(f Folder) publicFolder {
	return publicFolder{
		ID:        f.ID,
		ParentID:  f.ParentID,
		Name:      f.Name,
		CreatedAt: f.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: f.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func browserJSON(b Browser) gin.H {
	folders := make([]publicFolder, 0, len(b.Folders))
	for _, f := range b.Folders {
		folders = append(folders, publicFrom(f))
	}
	crumb := make([]publicFolder, 0, len(b.Breadcrumb))
	for _, f := range b.Breadcrumb {
		crumb = append(crumb, publicFrom(f))
	}
	files := make([]publicFile, 0, len(b.Files))
	for _, f := range b.Files {
		files = append(files, publicFile{
			ID:        f.ID,
			Name:      f.Name,
			MimeType:  f.MimeType,
			SizeBytes: f.SizeBytes,
			UpdatedAt: f.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	var current any
	if b.Folder != nil {
		p := publicFrom(*b.Folder)
		current = p
	} else {
		current = nil
	}
	return gin.H{
		"folder":     current,
		"breadcrumb": crumb,
		"folders":    folders,
		"files":      files,
	}
}
