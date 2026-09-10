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
		before, err := ParseTimelineCursor(c.Query("before"))
		if err != nil {
			httpx.Validation(c)
			return
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
		resp := gin.H{"groups": publicFilteredGroups(out.Groups)}
		if out.NextBefore != "" {
			resp["nextBefore"] = out.NextBefore
		}
		c.JSON(http.StatusOK, resp)
	})...)
}

func publicFilteredGroups(groups []TimelineGroup) []gin.H {
	out := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		items := make([]gin.H, 0, len(group.Items))
		for _, item := range group.Items {
			public := gin.H{
				"id":         item.ID,
				"name":       item.Name,
				"mimeType":   item.MimeType,
				"sizeBytes":  item.SizeBytes,
				"createdAt":  item.CreatedAt.UTC().Format(time.RFC3339Nano),
				"isFavorite": item.IsFavorite,
			}
			if item.ThumbnailURL != "" {
				public["thumbnailUrl"] = item.ThumbnailURL
			}
			items = append(items, public)
		}
		out = append(out, gin.H{"date": group.Date, "items": items})
	}
	return out
}
