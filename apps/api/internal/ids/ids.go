// Package ids generates ULID identifiers shared across modules.
package ids

import "github.com/oklog/ulid/v2"

// New returns a new ULID string (lexicographically sortable by time).
func New() string {
	return ulid.Make().String()
}
