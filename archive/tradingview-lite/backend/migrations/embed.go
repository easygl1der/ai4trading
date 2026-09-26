package migrations

import "embed"

// FS contains the versioned schema used by the application at startup.
//
//go:embed *.sql
var FS embed.FS
