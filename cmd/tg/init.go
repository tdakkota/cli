package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

func initFlags() []cli.Flag {
	return append([]cli.Flag{
		&cli.IntFlag{
			Name:        "app-id",
			Value:       telegram.TestAppID,
			Usage:       "telegram app ID",
			EnvVars:     []string{"APP_ID"},
			DefaultText: "Telegram's test APP_ID",
		},
		&cli.StringFlag{
			Name:        "app-hash",
			Value:       telegram.TestAppHash,
			Usage:       "telegram app hash",
			EnvVars:     []string{"APP_HASH"},
			DefaultText: "Telegram's test APP_HASH",
		},
	}, commonFlags()...)
}

func writeConfig(cfgPath string, cfg Config) error {
	buf := new(bytes.Buffer)
	e := yaml.NewEncoder(buf)
	e.SetIndent(2)

	if err := e.Encode(cfg); err != nil {
		return xerrors.Errorf("encode: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
		return xerrors.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(cfgPath, buf.Bytes(), 0600); err != nil {
		return xerrors.Errorf("write config: %w", err)
	}

	return nil
}

func genericInit(
	c *cli.Context,
	sessionName string,
	cb func(ctx context.Context, client *telegram.Client) error,
) error {
	ctx := c.Context
	sampleCfg := Config{
		Version: 1,
		AppID:   c.Int("app-id"),
		AppHash: c.String("app-hash"),
	}

	cfgPath := c.String("config")
	if cfgPath == "" {
		return xerrors.Errorf("no config path provided")
	}

	sessionPath := filepath.Join(filepath.Dir(cfgPath), sessionName)
	switch _, err := os.Stat(sessionPath); {
	case err == nil:
		return xerrors.Errorf("session %q already exist", sessionPath)
	case !os.IsNotExist(err):
		return xerrors.Errorf("stat: %w", err)
	}

	opts := telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionPath,
		},
		NoUpdates: true,
	}
	if err := propagateOptions(c, cfgPath, &opts); err != nil {
		return err
	}

	client := telegram.NewClient(sampleCfg.AppID, sampleCfg.AppHash, opts)
	if err := client.Run(ctx, func(ctx context.Context) error {
		return cb(ctx, client)
	}); err != nil {
		return multierr.Combine(err, os.Remove(sessionPath))
	}

	sampleCfg.Session = sessionPath
	if err := writeConfig(cfgPath, sampleCfg); err != nil {
		return err
	}
	fmt.Println("Wrote config to", cfgPath)
	return nil
}
