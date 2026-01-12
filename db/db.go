package db

import "embed"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

//go:embed fixtures/*.yml
var FixtureFS embed.FS
