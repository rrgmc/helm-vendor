package cmd

import (
	"fmt"

	"github.com/rrgmc/helm-vendor/internal/config"
)

type Cmd struct {
	cfg        config.Config
	outputRoot string
}

func New(cfg config.Config, options ...Option) (*Cmd, error) {
	ret := &Cmd{
		cfg: cfg,
	}
	for _, option := range options {
		option(ret)
	}
	return ret, nil
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
		cmd.outputRoot = outputRoot
	}
}

type Option func(*Cmd)
