package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/webitel/webitel-ai-hub/config"
)

// CMD returns the long-running server command.
func CMD() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "Run the service",
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
	cfg, err := config.LoadServerConfig()
	if err != nil {
		return err
	}

	app := NewApp(cfg)
	if err := app.Start(c.Context); err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")

	return app.Stop(context.Background())
}
