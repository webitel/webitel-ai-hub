package migrations

import (
	"strings"
	"testing"
)

func TestSchemaVersionTable(t *testing.T) {
	if SchemaVersionTable != "webitel_ai_hub_schema_version" {
		t.Fatalf("SchemaVersionTable = %q", SchemaVersionTable)
	}
}

func TestInitMigrationCreatesSchema(t *testing.T) {
	body, err := EmbedMigrations.ReadFile("20261019120000_init_ai_hub_schema.sql")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	sql := string(body)
	for _, want := range []string{
		"-- +goose Up",
		"CREATE SCHEMA ai_hub;",
		"-- +goose Down",
		"DROP SCHEMA IF EXISTS ai_hub CASCADE;",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration missing %q\n%s", want, sql)
		}
	}
}
