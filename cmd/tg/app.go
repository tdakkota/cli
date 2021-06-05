package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type app struct {
	cfg  Config
	log  *zap.Logger
	opts telegram.Options

	debugInvoker bool
}

func newApp() *app {
	// TODO(ernado): We need to log somewhere until configured?
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Level.SetLevel(zap.WarnLevel)

	defaultLog, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}

	return &app{
		log: defaultLog,
		opts: telegram.Options{
			NoUpdates: true,
		},
	}
}

type SessionNotAuthorizedError struct {
	Session string
}

func (s SessionNotAuthorizedError) Error() string {
	return fmt.Sprintf("session %q not authorizored", s.Session)
}

func (p *app) run(ctx context.Context, f func(ctx context.Context, api *tg.Client) error) error {
	client := telegram.NewClient(p.cfg.AppID, p.cfg.AppHash, p.opts)

	return client.Run(ctx, func(ctx context.Context) error {
		s, err := client.Auth().Status(ctx)
		if err != nil {
			return xerrors.Errorf("check auth status: %w", err)
		}
		if !s.Authorized {
			return &SessionNotAuthorizedError{
				Session: p.cfg.Session,
			}
		}

		return f(ctx, client.API())
	})
}

func (p *app) Before(c *cli.Context) error {
	cfgPath := c.String("config")
	if cfgPath == "" {
		return xerrors.New("no config path provided")
	}

	if err := propagateOptions(c, cfgPath, &p.opts); err != nil {
		return err
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return xerrors.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &p.cfg); err != nil {
		return xerrors.Errorf("parse config: %w", err)
	}

	p.opts.SessionStorage = &session.FileStorage{
		Path: filepath.Join(filepath.Dir(cfgPath), p.cfg.Session),
	}
	return nil
}
