package assets

import (
	"embed"
	"io/fs"
	"log"
)

// embed all static files
//
//go:embed all:static
var staticFS embed.FS

// embed all template files
//
//go:embed templates/*.tpl
var templateFS embed.FS

// embed all migration files
//
//go:embed migrations
var migrationFS embed.FS

// StaticFS returns the static files filesystem and "static" stripped from paths
func StaticFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal("failed to create static files sub filesystem:", err)
	}
	return sub
}

// TemplateFS returns the template files filesystem
func TemplateFS() fs.FS {
	return templateFS
}

// MigrationFS returns the migration files filesystem
func MigrationFS() fs.FS {
	return migrationFS
}
