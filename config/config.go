// Package config loads process settings from flags, the environment and an optional file.
package config

import (
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/pkg/errors"
)

const (
	defaultHealthStartTimeoutSec = 60
	defaultHealthStopTimeoutSec  = 12

	// pflagStringSliceDefValue is copied by appconfig from an unset StringSlice
	// DefValue into viper, so Load can yield []string{"[]"} without this filter.
	pflagStringSliceDefValue = "[]"
)

// Config is the subset of process settings this binary reads.
type Config struct {
	Log      appconfig.Log  `mapstructure:"log"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Health   HealthConfig   `mapstructure:"health"`
}

// PostgresConfig holds pool settings plus optional read-replica DSNs.
// Field names mirror appconfig.Postgres so shared flags/env still bind.
type PostgresConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	StandbyDSN      []string      `mapstructure:"standby_dsn"`
}

// StandbyDSNs returns trimmed non-empty standby DSNs.
func (p PostgresConfig) StandbyDSNs() []string {
	if len(p.StandbyDSN) == 0 {
		return nil
	}

	out := make([]string, 0, len(p.StandbyDSN))
	for _, dsn := range p.StandbyDSN {
		s := strings.TrimSpace(dsn)
		if s == "" || s == pflagStringSliceDefValue {
			continue
		}

		out = append(out, s)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// HealthConfig exposes the health probes and sd_notify timings.
type HealthConfig struct {
	// Address is the listen address of /livez, /readyz and /healthz. Empty disables them.
	Address string `mapstructure:"address"`
	// StartTimeout is seconds before sd_notify reports READY=1 anyway.
	StartTimeout int `mapstructure:"start_timeout"`
	// StopTimeout is the drain budget on shutdown, in seconds.
	StopTimeout int `mapstructure:"stop_timeout"`
}

// StartTimeoutDuration returns StartTimeout as a duration.
func (h HealthConfig) StartTimeoutDuration() time.Duration {
	return time.Duration(h.StartTimeout) * time.Second
}

// StopTimeoutDuration returns StopTimeout as a duration.
func (h HealthConfig) StopTimeoutDuration() time.Duration {
	return time.Duration(h.StopTimeout) * time.Second
}

// LoadServerConfig loads logging, Postgres and the health settings.
func LoadServerConfig() (*Config, error) {
	cfg, err := load(appconfig.Sections{Log: true, Postgres: true}, registerServerExtraFlags)
	if err != nil {
		return nil, err
	}

	if err := cfg.validateServer(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadMigrateConfig loads the logging and Postgres settings the migrate command needs.
func LoadMigrateConfig() (*Config, error) {
	cfg, err := load(appconfig.Sections{Log: true, Postgres: true}, nil)
	if err != nil {
		return nil, err
	}

	if err := cfg.requireDSN(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func load(sections appconfig.Sections, extra func(*pflag.FlagSet)) (*Config, error) {
	loader := appconfig.NewLoader(sections)
	loader.RegisterFlags(pflag.CommandLine)

	if extra != nil {
		extra(pflag.CommandLine)
	}

	pflag.Parse()

	return finishLoad(loader, pflag.CommandLine)
}

func finishLoad(loader *appconfig.Loader, fs *pflag.FlagSet) (*Config, error) {
	cfg := &Config{}
	if err := loader.Load(fs, cfg); err != nil {
		return nil, err
	}

	cfg.Postgres.StandbyDSN = cfg.Postgres.StandbyDSNs()

	return cfg, nil
}

func registerServerExtraFlags(fs *pflag.FlagSet) {
	registerPostgresStandbyFlags(fs)
	registerHealthFlags(fs)
}

func registerPostgresStandbyFlags(fs *pflag.FlagSet) {
	fs.StringSlice("postgres.standby_dsn", nil, "optional PostgreSQL read-replica DSNs")
}

func registerHealthFlags(fs *pflag.FlagSet) {
	fs.String("health.address", "", "listen address of the health probes (empty disables them)")
	fs.Int("health.start_timeout", defaultHealthStartTimeoutSec,
		"seconds before sd_notify reports READY=1 anyway; must stay under TimeoutStartSec")
	fs.Int("health.stop_timeout", defaultHealthStopTimeoutSec,
		"seconds budget for readiness drain on shutdown; must fit inside TimeoutStopSec")
}

func (c *Config) validateServer() error {
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}

	if c.Health.StartTimeout <= 0 {
		c.Health.StartTimeout = defaultHealthStartTimeoutSec
	}

	if c.Health.StopTimeout <= 0 {
		c.Health.StopTimeout = defaultHealthStopTimeoutSec
	}

	return c.requireDSN()
}

func (c *Config) requireDSN() error {
	if c.Postgres.DSN == "" {
		return errors.New("config: postgres.dsn is required (use --postgres.dsn or POSTGRES_DSN env)")
	}

	return nil
}
