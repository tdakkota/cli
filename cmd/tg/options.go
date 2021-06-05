package main

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/urfave/cli/v2"

	"github.com/gotd/cli/internal/pretty"
)

func propagateOptions(c *cli.Context, cfgPath string, opts *telegram.Options) error {
	if c.Bool("test") {
		opts.DCList = dcs.Staging()
	}
	if c.Bool("debug-invoker") {
		opts.Middlewares = append(opts.Middlewares, telegram.MiddlewareFunc(pretty.Pretty))
	}

	return nil
}
