package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/file"
)

type Cmd struct {
	cfg            config.Config
	outputRootPath string
	outputRoot     *os.Root
}

func New(cfg config.Config, options ...Option) (*Cmd, error) {
	ret := &Cmd{
		cfg: cfg,
	}
	for _, option := range options {
		option(ret)
	}
	if cfg.OutputPath != "" {
		ret.outputRootPath = filepath.Join(ret.outputRootPath, cfg.OutputPath)
	}

	_ = os.MkdirAll(ret.outputRootPath, os.ModePerm)

	var err error
	ret.outputRoot, err = os.OpenRoot(ret.outputRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open output path: %w", err)
	}
	return ret, nil
}

func (c *Cmd) Close() {
	_ = c.outputRoot.Close()
}

func NewFromFile(configFile string, options ...Option) (*Cmd, error) {
	cfg, err := config.LoadFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return New(cfg, options...)
}

func WithOutputRoot(outputRoot string) Option {
	return func(cmd *Cmd) {
		cmd.outputRootPath = outputRoot
	}
}

type Option func(*Cmd)

func (c *Cmd) openChartRoot(chartConfig config.Chart) (*os.Root, error) {
	r, err := c.outputRoot.OpenRoot(filepath.Clean(chartConfig.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to open chart path: %w", err)
	}
	return r, nil
}

func (c *Cmd) chartRootExists(chartConfig config.Chart) bool {
	return file.Exists(c.outputRoot, filepath.Clean(chartConfig.Path))
}

func (c *Cmd) createChartRoot(chartConfig config.Chart) error {
	return c.outputRoot.MkdirAll(filepath.Clean(chartConfig.Path), os.ModePerm)
}

func (c *Cmd) chartRootFileExists(chartConfig config.Chart) bool {
	return file.Exists(c.outputRoot, filepath.Join(filepath.Clean(chartConfig.Path), "Chart.yaml"))
}
