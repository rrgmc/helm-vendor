package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/helm"
)

func Dependency(ctx context.Context, path string, name string, version string, allVersions bool, outputPath string) error {
	currentChartFilename := filepath.Join(path, "Chart.yaml")

	chart, err := helm.LoadHelmChartVersionFilename(currentChartFilename)
	if err != nil {
		return fmt.Errorf("error loading chart file %s: %w\n", currentChartFilename, err)
	}

	for _, dependency := range chart.Dependencies {
		isName := dependency.Name == name || dependency.Alias == name
		if name != "" && !isName {
			continue
		}

		currentVersion := dependency.Version
		var currentOutputPath string
		if isName && version != "" {
			currentVersion = version
		}
		if isName {
			currentOutputPath = outputPath
		}

		if !isName && dependency.Repository == "" {
			continue
		}

		err = Download(ctx, dependency.Repository, dependency.Name, currentVersion, allVersions, currentOutputPath)
		if err != nil {
			return err
		}
	}

	return nil
}
