// Package migration is a shell to embed migration scripts
package migration

import (
	"embed"
)

// EmbeddedMigrations exposes migration scripts as an embedded filesystem
//go:embed *.sql
var EmbeddedMigrations embed.FS
