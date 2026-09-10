package photo

import (
	"net/http"
	"strconv"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

// RegisterFilteredTimelineRoute adds the typed timeline endpoint without changing
// the existing /photos/timeline contract used by current web/mobile clients.
func RegisterFilteredTimelineRoute(g *gin.RouterGroup, svc *Service, middleware ...gin.HandlerFunc) {
	g.GET("/photos/timeline/filter", append(middleware, func(c *gin.Context) {
		userID, ok := userIDFrom(c)
		if !ok {
			httpx.Error(c, apperr.Unauthorized)
			return
		}
		mediaType := c.Query("type")
		if mediaType != "image" && mediaType != "video" {
			httpx.Validation(c)
			return
		}
		var before *time.Time
		if raw := c.Query("before"); raw != "" {
			parsed, err := time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				httpx.Validation(c)
				return
			}
			before = &parsed
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
		out, err := svc.TimelineByType(c.Request.Context(), userID, before, limit, mediaType)
		if err != nil {
			httpx.Error(c, err)
			return
		}
		resp := gin.H{"groups": publicGroups(out.Groups)}
		if out.NextBefore != "" {
			resp["nextBefore"] = out.NextBefore
		}
		c.JSON(http.StatusOK, resp)
	})...)
}
