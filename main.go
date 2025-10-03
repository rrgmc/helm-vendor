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
		Usage: "Helm vendoring utilities",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-file",
				Aliases: []string{"c"},
				Value:   "helm-vendor.yaml",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "info",
				Usage: "Show chart information and versioning",
				Action: func(ctx context.Context, command *cli.Command) error {
					c, err := newCmd(command)
					if err != nil {
						return err
					}
					defer c.Close()

					if command.NArg() < 1 {
						return c.InfoAll(ctx)
					}

					return c.Info(ctx, command.Args().First())
				},
			},
			{
				Name:  "fetch",
				Usage: "Fetch new charts. If the chart was already fetched, use the 'upgrade' command",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "fetch all pending charts",
						Value:   false,
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					c, err := newCmd(command)
					if err != nil {
						return err
					}
					defer c.Close()

					if command.Bool("all") {
						return c.FetchAll(ctx)
					}

					if command.NArg() < 1 {
						return errors.New("path name is required")
					}
					var version string
					if command.NArg() > 1 {
						version = command.Args().Get(1)
					}

					return c.Fetch(ctx, command.Args().First(), version)
				},
			},
			{
				Name:  "upgrade",
				Usage: "Upgrade a chart to a new version",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "ignore-current",
						Usage: "ignore current release (just unpack the new version over it)",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "apply-patch",
						Usage: "create a diff of the local version and the chart of the same version, and patch the new version with it",
						Value: false,
					},
				},
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
					defer c.Close()

					return c.Upgrade(ctx, command.Args().First(), version, command.Bool("ignore-current"), command.Bool("apply-patch"))
				},
			},
		},
	}

	return commands.Run(ctx, os.Args)
}

func newCmd(command *cli.Command, options ...cmd.Option) (*cmd.Cmd, error) {
	cfgPath, err := filepath.Abs(command.String("config-file"))
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute config path: %w", err)
	}
	outputRoot := filepath.Dir(cfgPath)

	options = append(options, cmd.WithOutputRoot(outputRoot))

	return cmd.NewFromFile(command.String("config-file"), options...)
}
