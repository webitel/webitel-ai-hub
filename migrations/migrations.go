// Package migrations embeds the goose SQL for schema ai_hub.
//
// The server process does not apply these files. The migrate command does.
// Goose records versions in public.webitel_ai_hub_schema_version, outside ai_hub,
// because that table has to exist before the schema migration runs.
package migrations

import "embed"

// SchemaVersionTable is the goose version table in the shared database.
const SchemaVersionTable = "webitel_ai_hub_schema_version"

//go:embed *.sql
var EmbedMigrations embed.FS
