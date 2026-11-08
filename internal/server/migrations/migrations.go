// Package migrations embeds the ordered SQL that owns the sonar server's
// schema. tronc/migrate runs it at boot.
package migrations

import "embed"

// FS holds the migration files at its root, the layout tronc/migrate expects.
//
//go:embed *.sql
var FS embed.FS
