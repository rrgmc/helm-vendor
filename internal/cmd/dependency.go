package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/rrgmc/helm-vendor/internal/helm"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
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
	values := chartutil.Values{}
	var valuesErr error

	chart, err := helm.LoadDir(path, func(name string, fi os.FileInfo) bool {
		if name == "values.yaml" {
			valueFile := filepath.Join(path, name)
			bytes, err := os.ReadFile(valueFile)
			if err != nil {
				valuesErr = err
				return false
			}
			if err := yaml.Unmarshal(bytes, &values); err != nil {
				valuesErr = fmt.Errorf("failed to parse %s: %w", valueFile, err)
				return false
			}

			return false
		}
		return true
	})
	if err != nil {
		return err
	}
	if valuesErr != nil {
		return valuesErr
	}

	emptyValues := chartutil.Values{}

	if err := chartutil.ProcessDependencies(chart, emptyValues); err != nil {
		return err
	}

	releaseOptions := chartutil.ReleaseOptions{
		Name:      chart.Metadata.Name,
		Namespace: "default",
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}

	valuesToRender, err := chartutil.ToRenderValues(chart, emptyValues, releaseOptions, nil)
	if err != nil {
		return err
	}

	spew.Dump(values)
	spew.Dump(valuesToRender)

	return nil
}
