package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/pflag"

	"github.com/webitel/webitel-go-kit/appconfig"
)

func testPostgres(dsn string) PostgresConfig {
	return PostgresConfig{DSN: dsn}
}

func TestValidateServerRequiresDSN(t *testing.T) {
	cfg := &Config{}

	if err := cfg.validateServer(); err == nil {
		t.Fatal("validateServer() error = nil, want missing dsn")
	}
}

func TestValidateServerDefaultsLogLevel(t *testing.T) {
	cfg := &Config{Postgres: testPostgres("postgres://localhost/webitel")}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want info", cfg.Log.Level)
	}
}

func TestValidateServerDefaultsHealthTimeouts(t *testing.T) {
	cfg := &Config{Postgres: testPostgres("postgres://localhost/webitel")}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Health.StartTimeout != defaultHealthStartTimeoutSec {
		t.Fatalf("StartTimeout = %d, want %d", cfg.Health.StartTimeout, defaultHealthStartTimeoutSec)
	}

	if cfg.Health.StopTimeout != defaultHealthStopTimeoutSec {
		t.Fatalf("StopTimeout = %d, want %d", cfg.Health.StopTimeout, defaultHealthStopTimeoutSec)
	}
}

func TestRequireDSNAcceptsValue(t *testing.T) {
	cfg := &Config{Postgres: testPostgres("postgres://localhost/webitel")}

	if err := cfg.requireDSN(); err != nil {
		t.Fatalf("requireDSN() error = %v", err)
	}
}

func TestStandbyDSNs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "nil"},
		{name: "empty slice", in: []string{}},
		{name: "blank only", in: []string{"", "  "}},
		{name: "pflag unset artifact", in: []string{pflagStringSliceDefValue}},
		{
			name: "trims and drops blanks",
			in:   []string{" postgres://r1/db ", "", "postgres://r2/db"},
			want: []string{"postgres://r1/db", "postgres://r2/db"},
		},
		{
			name: "drops artifact beside real dsn",
			in:   []string{pflagStringSliceDefValue, "postgres://r1/db"},
			want: []string{"postgres://r1/db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PostgresConfig{StandbyDSN: tt.in}.StandbyDSNs()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("StandbyDSNs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestStandbyDSNLoadUnset(t *testing.T) {
	cfg := loadStandbyParsed(t, nil, nil)

	if got := cfg.Postgres.StandbyDSN; got != nil {
		t.Fatalf("StandbyDSN = %#v, want nil", got)
	}

	if got := cfg.Postgres.StandbyDSNs(); got != nil {
		t.Fatalf("StandbyDSNs() = %#v, want nil", got)
	}
}

func TestStandbyDSNLoadFromFlag(t *testing.T) {
	cfg := loadStandbyParsed(t, []string{
		"--postgres.standby_dsn=postgres://r1/db",
		"--postgres.standby_dsn=postgres://r2/db",
	}, nil)

	want := []string{"postgres://r1/db", "postgres://r2/db"}
	if got := cfg.Postgres.StandbyDSN; !slices.Equal(got, want) {
		t.Fatalf("StandbyDSN = %#v, want %#v", got, want)
	}
}

func TestStandbyDSNLoadFromEnv(t *testing.T) {
	cfg := loadStandbyParsed(t, nil, map[string]string{
		"POSTGRES_STANDBY_DSN": "postgres://e1/db,postgres://e2/db",
	})

	want := []string{"postgres://e1/db", "postgres://e2/db"}
	if got := cfg.Postgres.StandbyDSN; !slices.Equal(got, want) {
		t.Fatalf("StandbyDSN = %#v, want %#v", got, want)
	}
}

func TestStandbyDSNLoadFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.yml")

	content := []byte("postgres:\n  dsn: postgres://p/db\n  standby_dsn:\n    - postgres://y1/db\n    - \"  postgres://y2/db  \"\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := loadStandbyParsed(t, []string{"--config_file=" + path}, nil)

	want := []string{"postgres://y1/db", "postgres://y2/db"}
	if got := cfg.Postgres.StandbyDSN; !slices.Equal(got, want) {
		t.Fatalf("StandbyDSN = %#v, want %#v", got, want)
	}
}

func loadStandbyParsed(t *testing.T, args []string, env map[string]string) *Config {
	t.Helper()

	for k, v := range env {
		t.Setenv(k, v)
	}

	fs := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
	loader := appconfig.NewLoader(appconfig.Sections{Log: true, Postgres: true})
	loader.RegisterFlags(fs)
	registerPostgresStandbyFlags(fs)

	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse(%v) error = %v", args, err)
	}

	cfg, err := finishLoad(loader, fs)
	if err != nil {
		t.Fatalf("finishLoad() error = %v", err)
	}

	return cfg
}
