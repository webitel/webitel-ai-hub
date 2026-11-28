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

func TestPlpgsqlMigrationsUseGooseStatementBlock(t *testing.T) {
	t.Parallel()

	files := []string{
		"20261125120000_ai_connection.sql",
		"20261125120100_ai_profile.sql",
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			body, err := EmbedMigrations.ReadFile(name)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			sql := string(body)
			if !strings.Contains(sql, "LANGUAGE plpgsql") {
				t.Fatalf("expected plpgsql function in %s", name)
			}

			for _, want := range []string{
				"-- +goose StatementBegin",
				"-- +goose StatementEnd",
			} {
				if !strings.Contains(sql, want) {
					t.Fatalf("plpgsql migration missing %q (goose splits on ; otherwise)", want)
				}
			}
		})
	}
}

func TestMigrationsHaveAuditColumns(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want []string
	}{
		{
			name: "20261125120000_ai_connection.sql",
			want: []string{
				"created_by",
				"updated_by",
				"timestamptz NOT NULL DEFAULT now()",
			},
		},
		{
			name: "20261125120100_ai_profile.sql",
			want: []string{"created_by", "updated_by"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := EmbedMigrations.ReadFile(tc.name)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			sql := string(body)
			for _, want := range tc.want {
				if !strings.Contains(sql, want) {
					t.Fatalf("migration missing %q", want)
				}
			}
		})
	}
}
