package photo

import (
	"net/http"
	"strconv"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g.GET("/photos/timeline", append(middleware, h.timeline)...)
	g.GET("/photos/albums", append(middleware, h.listAlbums)...)
	g.POST("/photos/albums", append(middleware, h.createAlbum)...)
	g.GET("/photos/albums/:id", append(middleware, h.getAlbum)...)
	g.PATCH("/photos/albums/:id", append(middleware, h.patchAlbum)...)
	g.DELETE("/photos/albums/:id", append(middleware, h.deleteAlbum)...)
	g.POST("/photos/albums/:id/items", append(middleware, h.addItems)...)
	g.DELETE("/photos/albums/:id/items/:fileId", append(middleware, h.removeItem)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

func parseULIDParam(c *gin.Context, name string) (string, bool) {
	return httpx.ParamULID(c, name)
}

func (h *Handler) timeline(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var before *time.Time
	if raw := c.Query("before"); raw != "" {
		t, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			httpx.Validation(c)
			return
		}
		before = &t
	}
	limit := DefaultTimelineLimit
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			httpx.Validation(c)
			return
		}
		limit = n
	}
	out, err := h.svc.Timeline(c.Request.Context(), userID, before, limit)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	resp := gin.H{"groups": publicGroups(out.Groups)}
	if out.NextBefore != "" {
		resp["nextBefore"] = out.NextBefore
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) listAlbums(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	items, err := h.svc.ListAlbums(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, a := range items {
		out = append(out, publicAlbum(a))
	}
	c.JSON(http.StatusOK, gin.H{"albums": out})
}

type createAlbumReq struct {
	Name string `json:"name"`
}

func (h *Handler) createAlbum(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req createAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	a, err := h.svc.CreateAlbum(c.Request.Context(), userID, req.Name)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, publicAlbum(a))
}

func (h *Handler) getAlbum(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	albumID, ok := parseULIDParam(c, "id")
	if !ok {
		return
	}
	detail, err := h.svc.GetAlbum(c.Request.Context(), userID, albumID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicAlbumDetail(detail))
}

type patchAlbumReq struct {
	Name string `json:"name"`
}

func (h *Handler) patchAlbum(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	albumID, ok := parseULIDParam(c, "id")
	if !ok {
		return
	}
	var req patchAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	a, err := h.svc.PatchAlbum(c.Request.Context(), userID, albumID, req.Name)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicAlbum(a))
}

func (h *Handler) deleteAlbum(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	albumID, ok := parseULIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteAlbum(c.Request.Context(), userID, albumID); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type addItemsReq struct {
	FileIDs []string `json:"fileIds"`
}

func (h *Handler) addItems(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	albumID, ok := parseULIDParam(c, "id")
	if !ok {
		return
	}
	var req addItemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	for _, fid := range req.FileIDs {
		if _, err := ulid.ParseStrict(fid); err != nil {
			httpx.Validation(c)
			return
		}
	}
	if err := h.svc.AddAlbumItems(c.Request.Context(), userID, albumID, req.FileIDs); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) removeItem(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	albumID, ok := parseULIDParam(c, "id")
	if !ok {
		return
	}
	fileID, ok := parseULIDParam(c, "fileId")
	if !ok {
		return
	}
	if err := h.svc.RemoveAlbumItem(c.Request.Context(), userID, albumID, fileID); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func publicGroups(groups []TimelineGroup) []gin.H {
	out := make([]gin.H, 0, len(groups))
	for _, g := range groups {
		items := make([]gin.H, 0, len(g.Items))
		for _, it := range g.Items {
			items = append(items, publicItem(it))
		}
		out = append(out, gin.H{"date": g.Date, "items": items})
	}
	return out
}

func publicItem(it TimelineItem) gin.H {
	return gin.H{
		"id":        it.ID,
		"name":      it.Name,
		"mimeType":  it.MimeType,
		"sizeBytes": it.SizeBytes,
		"createdAt": it.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func publicAlbum(a Album) gin.H {
	return gin.H{
		"id":        a.ID,
		"name":      a.Name,
		"createdAt": a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updatedAt": a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func publicAlbumDetail(d AlbumDetail) gin.H {
	items := make([]gin.H, 0, len(d.Items))
	for _, it := range d.Items {
		items = append(items, publicItem(it))
	}
	resp := publicAlbum(d.Album)
	resp["items"] = items
	return resp
}
