package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/helm"
)

func Dependency(ctx context.Context, path string) error {
	currentChartFilename := filepath.Join(path, "Chart.yaml")

	chart, err := helm.LoadHelmChartVersionFilename(currentChartFilename)
	if err != nil {
		return fmt.Errorf("error loading chart file %s: %w\n", currentChartFilename, err)
	}

	for _, dependency := range chart.Dependencies {
		if dependency.Repository == "" {
			continue
		}
		err = Download(ctx, dependency.Repository, dependency.Name, dependency.Version, "")
		if err != nil {
			return err
		}
	}

	return nil
}
