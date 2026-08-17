package httpx

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

func ParamULID(c *gin.Context, name string) (string, bool) {
	id := c.Param(name)
	if _, err := ulid.ParseStrict(id); err != nil {
		Validation(c)
		return "", false
	}
	return id, true
}

func QueryULID(c *gin.Context, name string) (*string, bool) {
	q := strings.TrimSpace(c.Query(name))
	if q == "" {
		return nil, true
	}
	if _, err := ulid.ParseStrict(q); err != nil {
		Validation(c)
		return nil, false
	}
	return &q, true
}
