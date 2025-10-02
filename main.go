package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/cmd"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	commands := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-file",
				Aliases: []string{"c"},
				Value:   "helm-vendor.yaml",
			},
			&cli.StringFlag{
				Name:  "output-root",
				Value: "",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "check",
				Action: func(ctx context.Context, command *cli.Command) error {
					if command.NArg() < 1 {
						return errors.New("path name is required")
					}

					c, err := newCmd(command)
					if err != nil {
						return err
					}

					return c.Check(ctx, command.Args().First())
				},
			},
			{
				Name: "fetch",
				Action: func(ctx context.Context, command *cli.Command) error {
					if command.NArg() < 1 {
						return errors.New("path name is required")
					}
					var version string
					if command.NArg() > 1 {
						version = command.Args().Get(1)
					}

					c, err := newCmd(command)
					if err != nil {
						return err
					}

					return c.Fetch(ctx, command.Args().First(), version)
				},
			},
			{
				Name: "upgrade",
				Action: func(ctx context.Context, command *cli.Command) error {
					if command.NArg() < 1 {
						return errors.New("path name is required")
					}
					var version string
					if command.NArg() > 1 {
						version = command.Args().Get(1)
					}

					c, err := newCmd(command)
					if err != nil {
						return err
					}

					return c.Upgrade(ctx, command.Args().First(), version, false, false)
				},
			},
			{
				Name: "version-check",
				Action: func(ctx context.Context, command *cli.Command) error {
					c, err := newCmd(command)
					if err != nil {
						return err
					}

					return c.VersionCheck(ctx)
				},
			},
		},
	}

	return commands.Run(ctx, os.Args)
}

func newCmd(command *cli.Command, options ...cmd.Option) (*cmd.Cmd, error) {
	outputRoot := command.String("output-root")
	if outputRoot == "" {
		cfgPath, err := filepath.Abs(command.String("config-file"))
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute config path: %w", err)
		}
		outputRoot = filepath.Base(cfgPath)
	}

	return cmd.NewFromFile(command.String("config-file"),
		cmd.WithOutputRoot(outputRoot))
}
