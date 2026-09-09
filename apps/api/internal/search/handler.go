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
	q := Query{Text: c.Query("q")}

	typeName, err := ParseType(c.Query("type"))
	if err != nil {
		httpx.Validation(c)
		return
	}
	q.Type = typeName

	folderID, ok := httpx.QueryULID(c, "folderId")
	if !ok {
		return
	}
	q.FolderID = folderID

	from, err := ParseDateOnly(c.Query("from"))
	if err != nil {
		httpx.Validation(c)
		return
	}
	q.From = from
	to, err := ParseDateOnly(c.Query("to"))
	if err != nil {
		httpx.Validation(c)
		return
	}
	q.To = to

	sort, err := ParseSort(c.Query("sort"))
	if err != nil {
		httpx.Validation(c)
		return
	}
	q.Sort = sort

	order, err := ParseOrder(c.Query("order"))
	if err != nil {
		httpx.Validation(c)
		return
	}
	q.Order = order

	q.Limit = DefaultLimit
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 || n > MaxLimit {
			httpx.Validation(c)
			return
		}
		q.Limit = n
	}

	out, err := h.svc.Search(c.Request.Context(), userID, q)
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
			"createdAt": f.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updatedAt": f.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return gin.H{"folders": folders, "files": files}
}
