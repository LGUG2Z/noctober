package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/marcus-crane/october/backend"
	"github.com/urfave/cli/v2"
)

func IsCLIInvokedExplicitly(args []string) bool {
	for k, v := range os.Args {
		if k == 1 && v == "cli" {
			return true
		}
	}
	return false
}

func Invoke(isPortable bool, version string, logger *slog.Logger) {
	app := &cli.App{
		Name:     "noctober cli",
		HelpName: "noctober cli",
		Version:  version,
		Authors: []*cli.Author{
			{
				Name:  "Jeezy",
				Email: "support@notado.app",
			},
		},
		Usage: "sync your kobo highlights to notado from your terminal",
		Commands: []*cli.Command{
			{
				Name:    "sync",
				Aliases: []string{"s"},
				Usage:   "sync kobo highlights to notado",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					b, err := backend.StartBackend(&ctx, version, isPortable, logger)
					if err != nil {
						return err
					}
					if b.Settings.NotadoToken == "" {
						return fmt.Errorf("no notado token was configured. please set this up using the gui as the cli does not support this yet")
					}
					kobos := b.DetectKobos()
					if len(kobos) == 0 {
						return fmt.Errorf("no kobo was found. have you plugged one in and accepted the connection request?")
					}
					if len(kobos) > 1 {
						return fmt.Errorf("cli only supports one connected kobo at a time")
					}
					if err := b.SelectKobo(kobos[0].MntPath); err != nil {
						return fmt.Errorf("an error occurred trying to connect to the kobo at %s", kobos[0].MntPath)
					}
					num, err := b.ForwardToNotado()
					if err != nil {
						return err
					}
					logger.Info("Successfully synced highlights to Notado",
						slog.Int("count", num),
					)
					return nil
				},
			},
		},
	}

	// We remove the cli command so that urfave/cli doesn't try to literally parse it
	// but the help text of the cli tool still shows the user `october cli` so they don't
	// get disoriented and know that we're juggling text under the hood
	var args []string

	for k, v := range os.Args {
		if k == 1 && v == "cli" {
			continue
		}
		args = append(args, v)
	}

	err := app.Run(args)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
