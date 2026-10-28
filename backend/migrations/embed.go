// Package migrations embeds the versioned SQL migration files so the binary
// can apply them at startup without external tooling.
package migrations

import "embed"

// FS holds every .sql migration, applied in lexical order.
//
//go:embed *.sql
var FS embed.FS
