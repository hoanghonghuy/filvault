package activity

import (
	"errors"
	"time"
)

var errInvalidBefore = errors.New("invalid before cursor")

// parseCursor parses the `before` cursor (RFC3339Nano timestamp).
func parseCursor(raw string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, errInvalidBefore
	}
	return parsed.UTC(), nil
}
