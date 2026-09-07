package chat

import (
	"context"
	"fmt"
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
	g.POST("/chat/direct-conversations", append(middleware, h.createDirectConversation)...)
	g.GET("/chat/conversations/:id/messages", append(middleware, h.listMessages)...)
	g.POST("/chat/conversations/:id/messages", append(middleware, h.createMessage)...)
	g.PATCH("/chat/conversations/:id/messages/:messageId", append(middleware, h.editMessage)...)
	g.DELETE("/chat/conversations/:id/messages/:messageId", append(middleware, h.removeMessage)...)
	g.GET("/chat/conversations/:id/attachments/:attachmentId/download", append(middleware, h.downloadAttachment)...)
	g.GET("/chat/events", append(middleware, h.events)...)
	g.GET("/chat/conversations/:id/messages/search", append(middleware, h.searchMessages)...)
	g.GET("/chat/conversations/:id/media", append(middleware, h.listMedia)...)
	g.POST("/chat/attachments/upload-sessions", append(middleware, h.createAttachmentSession)...)
	g.POST("/chat/attachments/:fileId/complete", append(middleware, h.completeAttachment)...)
	g.POST("/chat/conversations/:id/call/token", append(middleware, h.getCallToken)...)
	g.POST("/chat/conversations/:id/call/signal", append(middleware, h.sendCallSignal)...)
	g.POST("/chat/conversations/:id/read", append(middleware, h.markAsRead)...)
	g.POST("/chat/conversations/:id/typing", append(middleware, h.sendTyping)...)
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

type createDirectConversationReq struct {
	RecipientEmail string `json:"recipientEmail"`
}

func (h *Handler) createDirectConversation(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req createDirectConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	out, err := h.svc.CreateDirectConversation(c.Request.Context(), userID, req.RecipientEmail)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicConversation(out))
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
	if c.Query("includePreview") == "true" {
		list, err := h.svc.ListConversationsWithPreview(c.Request.Context(), userID)
		if err != nil {
			httpx.Error(c, err)
			return
		}
		out := make([]gin.H, 0, len(list))
		for _, view := range list {
			out = append(out, publicConversationView(view))
		}
		c.JSON(http.StatusOK, gin.H{"conversations": out})
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
	Body            string   `json:"body"`
	FileIDs         []string `json:"fileIds"`
	ClientMessageID string   `json:"clientMessageId"`
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
	out, err := h.svc.CreateMessage(c.Request.Context(), userID, conversationID, req.Body, req.FileIDs, req.ClientMessageID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if out.Idempotent {
		c.JSON(http.StatusOK, publicMessage(out))
		return
	}
	c.JSON(http.StatusCreated, publicMessage(out))
}

type editMessageReq struct {
	Body string `json:"body"`
}

func (h *Handler) editMessage(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	messageID, ok := httpx.ParamULID(c, "messageId")
	if !ok {
		return
	}
	var req editMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	out, err := h.svc.EditMessage(c.Request.Context(), userID, conversationID, messageID, req.Body)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, publicMessage(out))
}

func (h *Handler) removeMessage(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	messageID, ok := httpx.ParamULID(c, "messageId")
	if !ok {
		return
	}
	if err := h.svc.RemoveMessage(c.Request.Context(), userID, conversationID, messageID); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) downloadAttachment(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	attachmentID, ok := httpx.ParamULID(c, "attachmentId")
	if !ok {
		return
	}
	url, err := h.svc.DownloadAttachment(c.Request.Context(), userID, conversationID, attachmentID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"downloadUrl": url.URL,
		"expiresAt":   url.ExpiresAt.UTC().Format(time.RFC3339Nano),
	})
}

func (h *Handler) events(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	after, err := strconv.ParseInt(c.Query("after"), 10, 64)
	if err != nil || after < 0 {
		after = 0
	}
	if raw := c.GetHeader("Last-Event-ID"); raw != "" {
		if headerAfter, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && headerAfter > after {
			after = headerAfter
		}
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Header("Access-Control-Allow-Headers", "Authorization, Last-Event-ID")
	c.Status(http.StatusOK)
	now := time.Now().UTC()
	_ = h.svc.HandleConnect(c.Request.Context(), userID, now)
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.svc.HandleDisconnect(disconnectCtx, userID, time.Now().UTC())
	}()
	heartbeat := time.NewTicker(20 * time.Second)
	poll := time.NewTicker(time.Second)
	defer heartbeat.Stop()
	defer poll.Stop()
	writeEvents := func() error {
		events, err := h.svc.repo.ListEventsAfter(c.Request.Context(), userID, after, 100)
		if err != nil {
			return err
		}
		for _, event := range events {
			if _, err := fmt.Fprintf(c.Writer, "id: %d\nevent: %s\ndata: %s\n\n", event.Sequence, event.Type, event.Payload); err != nil {
				return err
			}
			after = event.Sequence
		}
		if len(events) > 0 {
			c.Writer.Flush()
		}
		return nil
	}
	if err := writeEvents(); err != nil {
		return
	}
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(c.Writer, ": heartbeat\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		case <-poll.C:
			if err := writeEvents(); err != nil {
				return
			}
		}
	}
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
	page, err := h.svc.ListMessagePage(c.Request.Context(), userID, conversationID, c.Query("before"), limit)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	out := make([]gin.H, 0, len(page.Messages))
	for _, m := range page.Messages {
		out = append(out, publicMessage(m))
	}
	resp := gin.H{"messages": out, "hasMore": page.HasMore}
	if page.NextBefore != "" {
		resp["nextBefore"] = page.NextBefore
	}
	c.JSON(http.StatusOK, resp)
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
	if out.Idempotent {
		c.JSON(http.StatusOK, publicMessage(out))
		return
	}
	c.JSON(http.StatusCreated, publicMessage(out))
}

func publicConversation(c Conversation) gin.H {
	out := gin.H{
		"id":            c.ID,
		"title":         c.Title,
		"type":          c.Type,
		"createdAt":     c.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updatedAt":     c.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"lastMessageAt": c.LastMessage.UTC().Format(time.RFC3339Nano),
	}
	if c.PeerID != "" {
		peer := gin.H{"id": c.PeerID, "name": c.PeerName, "email": c.PeerEmail}
		if c.PeerLastSeenAt != nil {
			peer["lastSeenAt"] = c.PeerLastSeenAt.UTC().Format(time.RFC3339Nano)
		}
		out["peer"] = peer
	}
	return out
}

func publicConversationView(view ConversationView) gin.H {
	out := publicConversation(view.Conversation)
	out["unreadCount"] = view.UnreadCount
	if view.PeerStatus != "" {
		out["peerStatus"] = view.PeerStatus
	}
	if view.PeerLastSeenAt != nil {
		out["peerLastSeenAt"] = view.PeerLastSeenAt.UTC().Format(time.RFC3339Nano)
	}
	if view.LastReadAt != nil {
		out["lastReadAt"] = view.LastReadAt.UTC().Format(time.RFC3339Nano)
	}
	if view.LastReadMessageID != "" {
		out["lastReadMessageId"] = view.LastReadMessageID
	}
	if view.PeerLastReadAt != nil {
		out["peerLastReadAt"] = view.PeerLastReadAt.UTC().Format(time.RFC3339Nano)
	}
	if view.PeerLastReadMessageID != "" {
		out["peerLastReadMessageId"] = view.PeerLastReadMessageID
	}
	if view.Preview != nil {
		out["preview"] = gin.H{
			"messageId":   view.Preview.MessageID,
			"body":        view.Preview.Body,
			"createdAt":   view.Preview.CreatedAt.UTC().Format(time.RFC3339Nano),
			"attachments": view.Preview.AttachmentIDs,
		}
	}
	return out
}

func publicMessage(m Message) gin.H {
	attachments := make([]gin.H, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		attachments = append(attachments, publicAttachment(a))
	}
	out := gin.H{
		"id":              m.ID,
		"conversationId":  m.ConversationID,
		"body":            m.Body,
		"senderId":        m.SenderID,
		"clientMessageId": m.ClientMessageID,
		"createdAt":       m.CreatedAt.UTC().Format(time.RFC3339Nano),
		"attachments":     attachments,
	}
	if m.EditedAt != nil {
		out["editedAt"] = m.EditedAt.UTC().Format(time.RFC3339Nano)
	}
	if m.RemovedAt != nil {
		out["removedAt"] = m.RemovedAt.UTC().Format(time.RFC3339Nano)
		out["body"] = ""
	}
	return out
}

func publicAttachment(a Attachment) gin.H {
	out := gin.H{
		"id":           a.ID,
		"fileId":       a.FileID,
		"originalName": a.OriginalName,
		"name":         a.Name,
		"mimeType":     a.MimeType,
		"sizeBytes":    a.SizeBytes,
		"createdAt":    a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"availability": a.Availability,
	}
	if a.ThumbnailURL != "" {
		out["thumbnailUrl"] = a.ThumbnailURL
	}
	return out
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

type sendCallSignalReq struct {
	Action  string `json:"action"`
	IsVideo bool   `json:"isVideo"`
}

func (h *Handler) getCallToken(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	res, err := h.svc.GetCallToken(c.Request.Context(), userID, conversationID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) sendCallSignal(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	var req sendCallSignalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if err := h.svc.SendCallSignal(c.Request.Context(), userID, conversationID, req.Action, req.IsVideo); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusOK)
}

type markAsReadReq struct {
	MessageID string `json:"messageId"`
}

func (h *Handler) markAsRead(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	var req markAsReadReq
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.MarkAsRead(c.Request.Context(), userID, conversationID, req.MessageID); err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type sendTypingReq struct {
	Typing bool `json:"typing"`
}

func (h *Handler) sendTyping(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	conversationID, ok := httpx.ParamULID(c, "id")
	if !ok {
		return
	}
	var req sendTypingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if err := h.svc.SendTyping(c.Request.Context(), userID, conversationID, req.Typing); err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

