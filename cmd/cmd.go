// Package cmd is the process entrypoint. It exposes the server and migrate commands.
package cmd

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/webitel/webitel-ai-hub/cmd/migrate"
	"github.com/webitel/webitel-ai-hub/cmd/server"
	"github.com/webitel/webitel-ai-hub/internal/model"
)

// Run executes the CLI using os.Args.
func Run() error {
	app := &cli.App{
		Name:  model.ServiceName,
		Usage: "AI Hub service for the Webitel platform",
		Commands: []*cli.Command{
			server.CMD(),
			migrate.CMD(),
		},
	}

	return app.Run(os.Args)
}
