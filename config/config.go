// Package config loads process settings from flags, the environment and an optional file.
package config

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/pkg/errors"
)

const (
	defaultHealthStartTimeoutSec = 60
	defaultHealthStopTimeoutSec  = 12
)

// Config is the subset of process settings this binary reads.
type Config struct {
	Log      appconfig.Log      `mapstructure:"log"`
	Postgres appconfig.Postgres `mapstructure:"postgres"`
	Health   HealthConfig       `mapstructure:"health"`
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
	cfg, err := load(appconfig.Sections{Log: true, Postgres: true}, registerHealthFlags)
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

func registerHealthFlags() {
	pflag.String("health.address", "", "listen address of the health probes (empty disables them)")
	pflag.Int("health.start_timeout", defaultHealthStartTimeoutSec,
		"seconds before sd_notify reports READY=1 anyway; must stay under TimeoutStartSec")
	pflag.Int("health.stop_timeout", defaultHealthStopTimeoutSec,
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
