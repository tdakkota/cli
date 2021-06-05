package main

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/message/styling"
)

// commonFlags returns common flags for all commands.
func commonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   defaultConfigPath(),
			Usage:   "config to use",
		},
		&cli.BoolFlag{
			Name:  "debug-invoker",
			Usage: "use pretty-printing debug invoker",
		},
		&cli.BoolFlag{
			Name:    "test",
			Aliases: []string{"staging"},
			Usage:   "connect to telegram test server",
		},
	}
}

// messageFlags returns common flags for send and upload.
func messageFlags() []cli.Flag {
	return append([]cli.Flag{
		&cli.BoolFlag{
			Name:  "html",
			Usage: "use HTML styling",
		},
		&cli.BoolFlag{
			Name:  "silent",
			Usage: "send this message silently (no notifications for the receivers)",
		},
		&cli.BoolFlag{
			Name:  "nowebpage",
			Usage: "disable generation of the webpage preview",
		},
		&cli.DurationFlag{
			Name:  "schedule",
			Usage: "scheduled message date for scheduled messages",
		},
	}, commonFlags()...)
}

func applyMessageFlags(
	c *cli.Context,
	r *message.RequestBuilder,
	msg string,
) (*message.Builder, []styling.StyledTextOption) {
	p := &r.Builder

	if c.Bool("silent") {
		p = p.Silent()
	}

	if c.Bool("nowebpage") {
		p = p.NoWebpage()
	}

	if d := c.Duration("schedule"); c.IsSet("schedule") {
		p = p.Schedule(time.Now().Add(d))
	}

	option := styling.Plain(msg)
	if c.Bool("html") {
		option = html.String(nil, msg)
	}

	return p, []styling.StyledTextOption{option}
}
