package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	"github.com/rrgmc/helm-vendor/internal/helm"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

func ValuesDiff(ctx context.Context, path string, valueFiles []string, showDiff, showEquals bool, ignoreKeys []string) error {
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

			values = trimNilValues(values)

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

	for _, valueFile := range valueFiles {
		currentMap := map[string]interface{}{}

		bytes, err := os.ReadFile(valueFile)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return fmt.Errorf("failed to parse %s: %w", valueFile, err)
		}
		// Merge with the previous map
		chartutil.CoalesceTables(values, currentMap)
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

	defaultValues := valuesToRender["Values"].(chartutil.Values)

	var depOptions []string
	for _, dep := range chart.Metadata.Dependencies {
		if dep.Repository == "" {
			continue
		}
		depName := dep.Name
		if dep.Alias != "" {
			depName = dep.Alias
		}
		depOptions = append(depOptions, depName)
		// fmt.Printf("Adding dependency %s [%s]\n", dep.Name, dep.Version)
	}

	mapIterate(values, func(path []string, value any) {
		if len(path) == 0 {
			return
		}
		if !slices.Contains(depOptions, path[0]) {
			return
		}

		pathName := strings.Join(path, ".")
		for _, ik := range ignoreKeys {
			if ik == pathName || strings.HasPrefix(pathName, ik+".") {
				return
			}
		}

		// pathOutput := strings.Join(path, ".")
		var pathOutput string
		for _, p := range path {
			pathOutput += fmt.Sprintf("[%s] ", p)
		}

		otherValue, exists := findRecursive(defaultValues, path)

		isEquals := exists && cmp.Equal(value, otherValue)

		if showDiff && !isEquals {
			if !exists {
				fmt.Printf("DIFF[NE]: %s = '%v' [NOTEXISTS]\n", pathOutput, value)
			} else {
				fmt.Printf("DIFF: %s = '%v' [was: '%v']\n", pathOutput, value, otherValue)
			}
		}
		if showEquals && isEquals {
			fmt.Printf("EQUALS: %s = '%v'\n", pathOutput, value)
		}
	})

	return nil
}

func ValuesRender(ctx context.Context, path string, valueFiles []string, excludeRootValues bool) error {
	chart, err := helm.LoadDir(path, func(name string, fi os.FileInfo) bool {
		if excludeRootValues && name == "values.yaml" {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	values := chartutil.Values{}

	for _, valueFile := range valueFiles {
		currentMap := map[string]interface{}{}

		bytes, err := os.ReadFile(valueFile)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return fmt.Errorf("failed to parse %s: %w", valueFile, err)
		}
		// Merge with the previous map
		chartutil.CoalesceTables(values, currentMap)
	}

	if err := chartutil.ProcessDependencies(chart, values); err != nil {
		return err
	}

	releaseOptions := chartutil.ReleaseOptions{
		Name:      chart.Metadata.Name,
		Namespace: "default",
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}

	valuesToRender, err := chartutil.ToRenderValues(chart, values, releaseOptions, nil)
	if err != nil {
		return err
	}

	renderedValues := valuesToRender["Values"].(chartutil.Values)

	return renderedValues.Encode(os.Stdout)
}

func mapIterate(m map[string]any, f func(path []string, value any)) {
	mapIteratePath(m, nil, f)
}

func mapIteratePath(m map[string]any, startPath []string, f func(path []string, value any)) {
	for k, v := range helm.MapSortedByKey(m) {
		currentPath := slices.Concat(startPath, []string{k})

		// If the value is another map, recurse
		if nextMap, isMap := v.(map[string]any); isMap {
			if len(nextMap) == 0 {
				f(currentPath, v)
			} else {
				mapIteratePath(nextMap, currentPath, f)
			}
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

func trimNilValues(vals map[string]interface{}) map[string]interface{} {
	valsCopy, err := copystructure.Copy(vals)
	if err != nil {
		return vals
	}
	valsCopyMap := valsCopy.(map[string]interface{})
	for key, val := range valsCopyMap {
		if val == nil {
			// Iterate over the values and remove nil keys
			delete(valsCopyMap, key)
		} else if istable(val) {
			// Recursively call into ourselves to remove keys from inner tables
			valsCopyMap[key] = trimNilValues(val.(map[string]interface{}))
		}
	}

	return valsCopyMap
}

// istable is a special-purpose function to see if the present thing matches the definition of a YAML table.
func istable(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}
