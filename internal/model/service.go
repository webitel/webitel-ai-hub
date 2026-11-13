// Package model holds process-wide constants that have no infrastructure imports.
package model

import "time"

const (
	// ServiceName is the binary, CLI, Consul and PostgreSQL application name.
	ServiceName = "webitel-ai-hub"
	// ServiceNamespace is the OpenTelemetry / platform namespace.
	ServiceNamespace = "webitel"
)

// Build metadata; ldflags overwrite these at release time.
var (
	Version        = "0.0.0"
	Commit         = "hash"
	CommitDate     = time.Now().String()
	Branch         = "branch"
	BuildTimestamp = ""
)
