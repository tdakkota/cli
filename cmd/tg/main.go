package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type Config struct {
	Version int    `yaml:"version"`
	AppID   int    `yaml:"app_id"`
	AppHash string `yaml:"app_hash"`
	Session string `yaml:"session"`
}

func defaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, "gotd", "gotd.cli.yaml")
}

func main() {
	p := newApp()
	app := &cli.App{
		Name:  "tg",
		Usage: "Telegram CLI",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Creates config and session",
				Description: `Command init creates config file and session at the given path.
Examples:
	tg init bot --app-id 10 --app-hash abcd --token token
	tg init user --app-id 10 --app-hash abcd --phone +123456789
`,
				Subcommands: []*cli.Command{
					{
						Name:  "bot",
						Usage: "Creates config file for bot",
						Description: `Command bot creates config file using bot token at the given path.
Example:
	tg init bot --app-id 10 --app-hash abcd --token token
`,
						Flags:  initBotFlags(),
						Action: initBotCmd,
					},
					{
						Name:  "user",
						Usage: "Creates config file for user",
						Description: `Command user creates config file at the given path.
Example:
	tg init user --app-id 10 --app-hash abcd --phone +123456789
`,
						Flags:  initUserFlags(),
						Action: initUserCmd,
					},
				},
			},
			{
				Name:      "send",
				Usage:     "Sends message to peer",
				ArgsUsage: "[message]",
				Flags:     p.sendFlags(),
				Action:    p.sendCmd,
			},
			{
				Name:      "upload",
				Aliases:   []string{"up"},
				Usage:     "Uploads file to peer",
				ArgsUsage: "[path]",
				Flags:     p.uploadFlags(),
				Action:    p.uploadCmd,
			},
		},
	}
	for _, cmd := range app.Commands {
		if cmd.Name != "init" {
			cmd.Before = p.Before
		}

		cmd.Flags = append(cmd.Flags, app.Flags...)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.RunContext(ctx, os.Args); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
