package chat

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
	g.GET("/chat/conversations", append(middleware, h.listConversations)...)
	g.POST("/chat/conversations", append(middleware, h.createConversation)...)
	g.GET("/chat/conversations/:id/messages", append(middleware, h.listMessages)...)
	g.POST("/chat/conversations/:id/messages", append(middleware, h.createMessage)...)
	g.GET("/chat/conversations/:id/messages/search", append(middleware, h.searchMessages)...)
	g.GET("/chat/conversations/:id/media", append(middleware, h.listMedia)...)
	g.POST("/chat/attachments/upload-sessions", append(middleware, h.createAttachmentSession)...)
	g.POST("/chat/attachments/:fileId/complete", append(middleware, h.completeAttachment)...)
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

type createConversationReq struct {
	Title string `json:"title"`
}

func (h *Handler) createConversation(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req createConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	out, err := h.svc.CreateConversation(c.Request.Context(), userID, req.Title)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, publicConversation(out))
}

func (h *Handler) listConversations(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	list, err := h.svc.ListConversations(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, conv := range list {
		out = append(out, publicConversation(conv))
	}
	c.JSON(http.StatusOK, gin.H{"conversations": out})
}

type createMessageReq struct {
	Body    string   `json:"body"`
	FileIDs []string `json:"fileIds"`
}

func (h *Handler) createMessage(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	var req createMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	out, err := h.svc.CreateMessage(c.Request.Context(), userID, conversationID, req.Body, req.FileIDs)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, publicMessage(out))
}

func (h *Handler) listMessages(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			httpx.Validation(c)
			return
		}
		limit = n
	}
	list, err := h.svc.ListMessages(c.Request.Context(), userID, conversationID, limit)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		out = append(out, publicMessage(m))
	}
	c.JSON(http.StatusOK, gin.H{"messages": out})
}

func (h *Handler) searchMessages(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	limit, ok := queryLimit(c)
	if !ok {
		return
	}
	list, err := h.svc.SearchMessages(c.Request.Context(), userID, conversationID, c.Query("q"), limit)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		out = append(out, publicMessage(m))
	}
	c.JSON(http.StatusOK, gin.H{"messages": out})
}

func (h *Handler) listMedia(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	limit, ok := queryLimit(c)
	if !ok {
		return
	}
	list, err := h.svc.ListMedia(c.Request.Context(), userID, conversationID, limit)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, a := range list {
		out = append(out, publicAttachment(a))
	}
	c.JSON(http.StatusOK, gin.H{"media": out})
}

type attachmentSessionReq struct {
	ConversationID string `json:"conversationId"`
	Name           string `json:"name"`
	Size           int64  `json:"size"`
	ContentType    string `json:"contentType"`
}

func (h *Handler) createAttachmentSession(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req attachmentSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if _, err := ulid.ParseStrict(req.ConversationID); err != nil {
		httpx.Validation(c)
		return
	}
	out, err := h.svc.CreateAttachmentSession(c.Request.Context(), userID, req.ConversationID, req.Name, req.ContentType, req.Size)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"fileId":    out.FileID,
		"uploadUrl": out.UploadURL,
		"expiresAt": out.ExpiresAt.UTC().Format(time.RFC3339Nano),
	})
}

type completeAttachmentReq struct {
	ConversationID string `json:"conversationId"`
	Body           string `json:"body"`
}

func (h *Handler) completeAttachment(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	fileID, ok := httpx.ParamULID(c, "fileId")
	if !ok {
		return
	}
	var req completeAttachmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if _, err := ulid.ParseStrict(req.ConversationID); err != nil {
		httpx.Validation(c)
		return
	}
	out, err := h.svc.CompleteAttachment(c.Request.Context(), userID, req.ConversationID, fileID, req.Body)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, publicMessage(out))
}

func publicConversation(c Conversation) gin.H {
	return gin.H{
		"id":        c.ID,
		"title":     c.Title,
		"createdAt": c.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updatedAt": c.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func publicMessage(m Message) gin.H {
	attachments := make([]gin.H, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		attachments = append(attachments, publicAttachment(a))
	}
	return gin.H{
		"id":             m.ID,
		"conversationId": m.ConversationID,
		"body":           m.Body,
		"createdAt":      m.CreatedAt.UTC().Format(time.RFC3339Nano),
		"attachments":    attachments,
	}
}

func publicAttachment(a Attachment) gin.H {
	return gin.H{
		"id":           a.ID,
		"fileId":       a.FileID,
		"originalName": a.OriginalName,
		"name":         a.Name,
		"mimeType":     a.MimeType,
		"sizeBytes":    a.SizeBytes,
		"createdAt":    a.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func queryLimit(c *gin.Context) (int, bool) {
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			httpx.Validation(c)
			return 0, false
		}
		limit = n
	}
	return limit, true
}
