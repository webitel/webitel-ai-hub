package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"github.com/webitel/webitel-ai-hub/cmd/server"
	"github.com/webitel/webitel-ai-hub/config"
	"github.com/webitel/webitel-ai-hub/migrations"
)

// CMD returns the migrate command. The server does not run it.
func CMD() *cli.Command {
	return &cli.Command{
		Name:    "migrate",
		Aliases: []string{"m"},
		Usage:   "Execute database migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config_file",
				Usage: "Path to the configuration file",
			},
		},
		Action: run,
	}
}

func run(c *cli.Context) error {
	cfg, err := config.LoadMigrateConfig()
	if err != nil {
		return err
	}

	app := fx.New(
		fx.Provide(
			func() *config.Config { return cfg },
			server.ProvideLogger,
		),
		fx.Invoke(func(loaded *config.Config, log *slog.Logger) error {
			return NewMigrator(loaded, log).Run(c.Context)
		}),
		fx.NopLogger,
	)

	return app.Start(c.Context)
}

type migrator struct {
	cfg *config.Config
	log *slog.Logger
}

// NewMigrator applies embedded goose migrations once.
func NewMigrator(cfg *config.Config, log *slog.Logger) *migrator {
	return &migrator{cfg: cfg, log: log}
}

// Run opens Postgres and applies every pending migration.
func (m *migrator) Run(ctx context.Context) error {
	db, err := openDB(m.cfg.Postgres.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetLogger(newLogger(m.log))
	goose.SetVerbose(true)

	provider, err := newProvider(db)
	if err != nil {
		return err
	}

	return m.apply(ctx, provider)
}

func openDB(dsn string) (*sql.DB, error) {
	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return stdlib.OpenDB(*conf.ConnConfig), nil
}

func newProvider(db *sql.DB) (*goose.Provider, error) {
	store, err := database.NewStore(database.DialectPostgres, migrations.SchemaVersionTable)
	if err != nil {
		return nil, err
	}

	return goose.NewProvider(goose.DialectCustom, db, migrations.EmbedMigrations, goose.WithStore(store))
}

func (m *migrator) apply(ctx context.Context, provider *goose.Provider) error {
	results, err := provider.Up(ctx)
	for _, result := range results {
		m.logResult(result)
	}

	return err
}

func (m *migrator) logResult(result *goose.MigrationResult) {
	if result == nil || result.Source == nil {
		return
	}

	if result.Error != nil {
		m.log.Error("unable to apply migration", "name", result.Source.Path, "err", result.Error)

		return
	}

	m.log.Info("applied migration", "name", result.Source.Path)
}

type migrateLogger struct {
	log *slog.Logger
}

func newLogger(log *slog.Logger) *migrateLogger {
	return &migrateLogger{log: log}
}

func (l *migrateLogger) Printf(format string, args ...any) {
	l.log.Info(fmt.Sprintf(format, args...))
}

func (l *migrateLogger) Fatalf(format string, args ...any) {
	l.log.Error(fmt.Sprintf(format, args...))
}
