// Package config loads process settings from flags, the environment and an optional file.
package config

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/pkg/errors"
)

const defaultHealthStartTimeoutSec = 60

// Config is the subset of process settings this binary reads.
type Config struct {
	Service  ServiceConfig      `mapstructure:"service"`
	Log      appconfig.Log      `mapstructure:"log"`
	Postgres appconfig.Postgres `mapstructure:"postgres"`
	Consul   appconfig.Consul   `mapstructure:"consul"`
	Health   HealthConfig       `mapstructure:"health"`
}

// ServiceConfig holds the gRPC listen address and connection TLS settings.
type ServiceConfig struct {
	Addr       string             `mapstructure:"addr"`
	Connection appconfig.GRPCConn `mapstructure:"conn"`
}

// HealthConfig exposes the health probes and sd_notify timings.
type HealthConfig struct {
	// Address is the listen address of /livez, /readyz and /healthz. Empty disables them.
	Address string `mapstructure:"address"`
	// StartTimeout is seconds before sd_notify reports READY=1 anyway.
	StartTimeout int `mapstructure:"start_timeout"`
	// StopTimeout is the drain budget on shutdown, in seconds.
	// Zero lets healthfx use DrainHold plus five seconds.
	StopTimeout int `mapstructure:"stop_timeout"`
}

func (h HealthConfig) StartTimeoutDuration() time.Duration {
	return time.Duration(h.StartTimeout) * time.Second
}

func (h HealthConfig) StopTimeoutDuration() time.Duration {
	return time.Duration(h.StopTimeout) * time.Second
}

// LoadServerConfig loads logging, Postgres, Consul, service and the health settings.
func LoadServerConfig() (*Config, error) {
	cfg, err := load(
		appconfig.Sections{Log: true, Postgres: true, Consul: true},
		registerServerFlags,
		appconfig.GRPCConnSecretPaths("service.conn")...,
	)
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

func load(sections appconfig.Sections, extra func(*pflag.FlagSet), secretPaths ...string) (*Config, error) {
	loader := appconfig.NewLoader(sections)
	loader.RegisterFlags(pflag.CommandLine)

	if len(secretPaths) > 0 {
		loader.MarkSecretPath(secretPaths...)
	}

	if extra != nil {
		extra(pflag.CommandLine)
	}

	pflag.Parse()

	cfg := &Config{}
	if err := loader.Load(pflag.CommandLine, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func registerServerFlags(fs *pflag.FlagSet) {
	fs.String("service.addr", "localhost:8080", "gRPC listen address")
	appconfig.RegisterGRPCConnFlags(fs, "service.conn", false)
	registerHealthFlags(fs)
}

func registerHealthFlags(fs *pflag.FlagSet) {
	fs.String("health.address", "", "listen address of the health probes (empty disables them)")
	fs.Int("health.start_timeout", defaultHealthStartTimeoutSec,
		"seconds before sd_notify reports READY=1 anyway; must stay under TimeoutStartSec")
	fs.Int("health.stop_timeout", 0,
		"seconds budget for readiness drain on shutdown; 0 uses the healthfx default (DrainHold + 5s)")
}

func (c *Config) validateServer() error {
	if c.Service.Addr == "" {
		return errors.New("config: service.addr is required")
	}

	if err := appconfig.ValidateGRPCConn("service.conn", c.Service.Connection); err != nil {
		return err
	}

	if c.Log.Level == "" {
		c.Log.Level = "info"
	}

	if c.Consul.Addr == "" {
		return errors.New("config: consul.addr is required")
	}

	if c.Health.StartTimeout <= 0 {
		c.Health.StartTimeout = defaultHealthStartTimeoutSec
	}

	return c.requireDSN()
}

func (c *Config) requireDSN() error {
	if c.Postgres.DSN == "" {
		return errors.New("config: postgres.dsn is required (use --postgres.dsn or POSTGRES_DSN env)")
	}

	return nil
}
