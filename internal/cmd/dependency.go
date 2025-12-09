package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/rrgmc/helm-vendor/internal/helm"
)

func Dependency(ctx context.Context, path string, allVersions bool) error {
	currentChartFilename := filepath.Join(path, "Chart.yaml")

	chart, err := helm.LoadHelmChartVersionFilename(currentChartFilename)
	if err != nil {
		return fmt.Errorf("error loading chart file %s: %w\n", currentChartFilename, err)
	}

	for _, dependency := range chart.Dependencies {
		if dependency.Repository == "" {
			continue
		}
		err = Download(ctx, dependency.Repository, dependency.Name, dependency.Version, allVersions, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func DependencyDiff(ctx context.Context, path string) error {
	chart, err := helm.LoadDir(path, func(name string, fi os.FileInfo) bool {
		if name == "values.yaml" {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	spew.Dump(chart)

	return nil
}
