package main

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

func initBotFlags() []cli.Flag {
	return append([]cli.Flag{
		&cli.StringFlag{
			Name:     "token",
			Required: true,
			Usage:    "telegram bot token",
			EnvVars:  []string{"BOT_TOKEN"},
		},
	}, initFlags()...)
}

func initBotCmd(c *cli.Context) error {
	token := c.String("token")
	if token == "" {
		return xerrors.Errorf("no token provided")
	}

	sessionName := fmt.Sprintf("gotd.session.%x.json", md5.Sum([]byte(token))) // #nosec
	return genericInit(c, sessionName, func(ctx context.Context, client *telegram.Client) error {
		s, err := client.Auth().Bot(ctx, token)
		if err != nil {
			return xerrors.Errorf("auth: %w", err)
		}

		if s.User != nil {
			if user, ok := s.User.AsNotEmpty(); ok {
				fmt.Printf("Logged as bot @%s\n", user.Username)
			}
		}
		return nil
	})
}
