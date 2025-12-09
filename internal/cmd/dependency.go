package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

	// valuesToRender, err := chartutil.ToRenderValues(chart, emptyValues, releaseOptions, nil)
	_, err = chartutil.ToRenderValues(chart, emptyValues, releaseOptions, nil)
	if err != nil {
		return err
	}

	mapIterate(values, func(path []string, value any) {
		fmt.Printf("%s : %v\n", strings.Join(path, "."), value)
	})

	// spew.Dump(values)
	// spew.Dump(valuesToRender)

	return nil
}

func mapIterate(m map[string]any, f func(path []string, value any)) {
	mapIteratePath(m, nil, f)
}

func mapIteratePath(m map[string]any, startPath []string, f func(path []string, value any)) {
	for k, v := range m {
		currentPath := slices.Concat(startPath, []string{k})

		// If the value is another map, recurse
		if nextMap, isMap := v.(map[string]any); isMap {
			mapIteratePath(nextMap, currentPath, f)
		} else {
			f(currentPath, v)
		}
	}
}

func findRecursive(data map[string]any, path []string) (interface{}, bool) {
	if len(path) == 0 {
		return data, true // Reached the target level
	}

	currentKey := path[0]
	val, ok := data[currentKey]
	if !ok {
		return nil, false // Key not found at this level
	}

	if len(path) == 1 {
		return val, true // Found the final value
	}

	// If the value is another map, recurse
	if nextMap, isMap := val.(map[string]any); isMap {
		return findRecursive(nextMap, path[1:])
	}

	return nil, false // Key found, but not a map for further recursion
}
