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
				Name:      "info",
				Usage:     "Show chart information and versioning",
				UsageText: "helm-vendor info [options] [path]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all-versions",
						Aliases: []string{"a"},
						Usage:   "shows all chart versions",
						Value:   false,
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					c, err := newCmd(command)
					if err != nil {
						return err
					}
					defer c.Close()

					if command.NArg() < 1 {
						return c.InfoAll(ctx)
					}

					return c.Info(ctx, command.Args().First(), command.Bool("all-versions"))
				},
			},
			{
				Name:      "fetch",
				Usage:     "Fetch new charts. If the chart was already fetched, use the 'upgrade' command",
				UsageText: "helm-vendor fetch [options] [path | --all]",
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
				Name:      "upgrade",
				Usage:     "Upgrade a chart to a new version",
				UsageText: "helm-vendor upgrade [options] path [version]",
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
					&cli.StringFlag{
						Name:  "latest-chart-path",
						Usage: "extract the latest chart in this path instead of a temporary",
					},
					&cli.StringFlag{
						Name:  "current-chart-path",
						Usage: "extract the current chart in this path instead of a temporary",
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

					return c.Upgrade(ctx, command.Args().First(), version, command.Bool("ignore-current"), command.Bool("apply-patch"),
						command.String("latest-chart-path"), command.String("current-chart-path"))
				},
			},
			{
				Name:      "download",
				Usage:     "Download a chart directly from a repository",
				UsageText: "helm-vendor download repoURL chartName [version]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all-versions",
						Aliases: []string{"a"},
						Usage:   "shows all chart versions",
						Value:   false,
					},
					&cli.StringFlag{
						Name:  "output-path",
						Usage: "output path to write the chart files. If empty, will only output the chart info",
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					if command.NArg() < 2 {
						return errors.New("repo URL and chart name is required")
					}
					var version string
					if command.NArg() > 2 {
						version = command.Args().Get(2)
					}
					return cmd.Download(ctx, command.Args().First(), command.Args().Get(1), version,
						command.Bool("all-versions"), command.String("output-path"))
				},
			},
			{
				Name:      "dependency",
				Usage:     "dependency",
				UsageText: "helm-vendor dependency",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all-versions",
						Aliases: []string{"a"},
						Usage:   "shows all chart versions",
						Value:   false,
					},
					&cli.StringFlag{
						Name:  "output-path",
						Usage: "output path to write the chart files. If empty, will only output the chart info (only if dependency-name is set)",
					},
					&cli.StringFlag{
						Name:  "name",
						Usage: "dependency name to check",
					},
					&cli.StringFlag{
						Name:  "version",
						Usage: "allows overriding the dependency version",
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					path, err := os.Getwd()
					if err != nil {
						return err
					}
					return cmd.Dependency(ctx, path, command.String("name"), command.String("version"),
						command.Bool("all-versions"), command.String("output-path"))
				},
			},
			{
				Name:      "values-diff",
				Usage:     "values-diff",
				UsageText: "helm-vendor values-diff",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "values",
						Aliases: []string{"f"},
						Usage:   "extra configuration values file name",
					},
					&cli.BoolFlag{
						Name:    "show-diff",
						Aliases: []string{"d"},
						Usage:   "shows diff",
						Value:   true,
					},
					&cli.BoolFlag{
						Name:    "show-equals",
						Aliases: []string{"e"},
						Usage:   "shows equals",
					},
					&cli.StringSliceFlag{
						Name:    "ignore-key",
						Aliases: []string{"i"},
						Usage:   "value keys to ignore",
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					path, err := os.Getwd()
					if err != nil {
						return err
					}
					return cmd.ValuesDiff(ctx, path, command.StringSlice("values"), command.Bool("show-diff"),
						command.Bool("show-equals"), command.StringSlice("ignore-key"))
				},
			},
			{
				Name:      "values-render",
				Usage:     "values-render",
				UsageText: "helm-vendor values-render",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "values",
						Aliases: []string{"f"},
						Usage:   "extra configuration values file name",
					},
					&cli.BoolFlag{
						Name:    "exclude-root-values",
						Aliases: []string{"e"},
						Usage:   "exclude root values file",
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					path, err := os.Getwd()
					if err != nil {
						return err
					}
					return cmd.ValuesRender(ctx, path, command.StringSlice("values"), command.Bool("exclude-root-values"))
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
