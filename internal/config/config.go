package config

import (
	"io"
	"os"

	"github.com/rrgmc/helm-vendor/internal/yaml"
)

type Config struct {
	Charts []Chart `yaml:"charts"`
}

type Chart struct {
	Path       string     `yaml:"path"`
	Repository Repository `yaml:"repository"`
	Name       string     `yaml:"name"`
	Files      Files      `yaml:"files"`
}

type Repository struct {
	URL string `yaml:"url"`
}

type Files struct {
	Ignore []string `yaml:"ignore"`
}

func Load(r io.Reader) (Config, error) {
	var cfg Config
	if err := yaml.Decode(r, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadFromFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	return Load(f)
}
