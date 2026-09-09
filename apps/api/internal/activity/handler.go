package activity

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

// Handler exposes GET /api/v1/activity (owner feed).
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(rg gin.IRouter, authMw ...gin.HandlerFunc) {
	v1 := rg.Group("/api/v1", authMw...)
	v1.GET("/activity", h.list)
}

type eventDTO struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	TargetName string `json:"targetName"`
	CreatedAt  string `json:"createdAt"`
}

type pageDTO struct {
	Events     []eventDTO `json:"events"`
	NextBefore *string    `json:"nextBefore"`
}

const (
	queryLimit  = "limit"
	queryBefore = "before"
)

func maxLimitMessage() string {
	return "limit must be between 1 and " + strconv.Itoa(MaxListLimit)
}

// ctxUserID mirrors auth.CtxUserID without importing auth (avoids a cycle
// once auth records activity events through this package).
const ctxUserID = "authUserID"

func (h *Handler) list(c *gin.Context) {
	ownerID, _ := c.Get(ctxUserID)
	id, _ := ownerID.(string)
	if id == "" {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	limit := DefaultListLimit
	if raw := c.Query(queryLimit); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > MaxListLimit {
			httpx.Error(c, apperr.New(http.StatusBadRequest, "VALIDATION_ERROR", maxLimitMessage()))
			return
		}
		limit = parsed
	}
	before := c.Query(queryBefore)

	page, err := h.svc.List(c.Request.Context(), id, before, limit)
	if err != nil {
		if errors.Is(err, errInvalidBefore) {
			httpx.Error(c, apperr.New(http.StatusBadRequest, "VALIDATION_ERROR", "before must be a valid cursor"))
			return
		}
		httpx.Error(c, err)
		return
	}

	dto := pageDTO{Events: make([]eventDTO, 0, len(page.Events))}
	for _, ev := range page.Events {
		dto.Events = append(dto.Events, eventDTO{
			ID:         ev.ID,
			Type:       ev.Type,
			TargetName: ev.TargetName,
			CreatedAt:  ev.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	if page.NextBefore != nil {
		s := page.NextBefore.UTC().Format(time.RFC3339Nano)
		dto.NextBefore = &s
	}
	c.JSON(http.StatusOK, dto)
}
