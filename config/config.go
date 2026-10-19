// Package config loads process settings from flags, the environment and an optional file.
package config

import (
	"github.com/spf13/pflag"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/pkg/errors"
)

// Config is the subset of process settings this binary reads.
type Config struct {
	Log      appconfig.Log      `mapstructure:"log"`
	Postgres appconfig.Postgres `mapstructure:"postgres"`
	Health   HealthConfig       `mapstructure:"health"`
}

// HealthConfig exposes the health probes.
type HealthConfig struct {
	// Addr is the listen address of /livez, /readyz and /healthz. Empty disables them.
	Addr string `mapstructure:"addr"`
}

// LoadServerConfig loads logging, Postgres and the health address.
func LoadServerConfig() (*Config, error) {
	cfg, err := load(appconfig.Sections{Log: true, Postgres: true}, registerHealthFlag)
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

func load(sections appconfig.Sections, extra func()) (*Config, error) {
	loader := appconfig.NewLoader(sections)
	loader.RegisterFlags(pflag.CommandLine)

	if extra != nil {
		extra()
	}

	pflag.Parse()

	cfg := &Config{}
	if err := loader.Load(pflag.CommandLine, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func registerHealthFlag() {
	pflag.String("health.addr", "", "listen address of the health probes (empty disables them)")
}

func (c *Config) validateServer() error {
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}

	return c.requireDSN()
}

func (c *Config) requireDSN() error {
	if c.Postgres.DSN == "" {
		return errors.New("config: postgres.dsn is required (use --postgres.dsn or POSTGRES_DSN env)")
	}

	return nil
}
