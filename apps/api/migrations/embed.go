package migrations

import "embed"

// FS holds numbered SQL files (*.up.sql / *.down.sql).
//
//go:embed *.sql
var FS embed.FS
