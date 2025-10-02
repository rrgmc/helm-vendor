package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/diff"
	"github.com/rrgmc/helm-vendor/internal/file"
	"github.com/rrgmc/helm-vendor/internal/helm"
)

func (c *Cmd) Upgrade(ctx context.Context, path string, version string, ignoreCurrent bool, applyPatch bool) error {
	for _, chartConfig := range c.cfg.Charts {
		if path == chartConfig.Path {
			return c.upgradeChart(ctx, chartConfig, version, ignoreCurrent, applyPatch)
		}
	}
	return fmt.Errorf("unknown path '%s'", path)
}

func (c *Cmd) upgradeChart(ctx context.Context, chartConfig config.Chart, version string, ignoreCurrent bool, applyPatch bool) error {
	chartOutputPath := c.buildChartPath(chartConfig)
	currentChartFilename := filepath.Join(chartOutputPath, "Chart.yaml")
	if !file.Exists(currentChartFilename) {
		return fmt.Errorf("chart not found in path '%s', use fetch to download an initial version", chartConfig.Path)
	}

	// load the Chart.yaml file for the current version
	currentChartVersionFile, err := helm.LoadHelmChartVersionFile(currentChartFilename)
	if err != nil {
		return fmt.Errorf("error loading current chart version file: %w", err)
	}

	repo, err := helm.LoadRepository(chartConfig.Repository.URL)
	if err != nil {
		return err
	}

	// download the new chart version
	latestChart, err := repo.GetChart(chartConfig.Name, version)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading new version of '%s' [%s - %s]\n", chartConfig.Path, latestChart.Chart().Name, latestChart.Chart().Version)

	latestChartFiles, err := latestChart.Download()
	if err != nil {
		return err
	}
	defer latestChartFiles.Close()

	chartFileIter := func(iter file.Iter) file.Iter {
		return file.IterFilter(iter, file.Filter{
			Ignore: chartConfig.Files.Ignore,
		})
	}

	diffBuilder := diff.NewBuilder(!ignoreCurrent)

	if !ignoreCurrent {
		fmt.Printf("Downloading source chart for local version [%s - %s]\n", currentChartVersionFile.Name, currentChartVersionFile.Version)

		sourceChart, err := repo.GetChart(currentChartVersionFile.Name, currentChartVersionFile.Version)
		if err != nil {
			return err
		}

		sourceChartFiles, err := sourceChart.Download()
		if err != nil {
			return err
		}
		defer sourceChartFiles.Close()

		// take diff of local code and chart code from the current version.
		for sourceChartFile, err := range chartFileIter(sourceChartFiles.Iter()) {
			if err != nil {
				return err
			}
			if sourceChartFile.Entry.IsDir() {
				continue
			}

			localFile := filepath.Join(chartOutputPath, sourceChartFile.Path)

			err = diffBuilder.Add(sourceChartFile.Path, sourceChartFile.Path, sourceChartFile.FullPath, localFile)
			if err != nil {
				return err
			}
		}

		// write diff
		if !diffBuilder.IsEmpty() {
			diffFilename, err := file.GenerateUniqueFilename(chartOutputPath,
				filepath.Clean(fmt.Sprintf("helm-vendor-%s-%s", chartConfig.Path, currentChartVersionFile.Version)),
				".diff")
			if err != nil {
				return fmt.Errorf("error generating unique diff filename: %w", err)
			}

			fmt.Printf("Writing diff file with changes between local and source chart\n")

			err = os.WriteFile(diffFilename, diffBuilder.Bytes(), os.ModePerm)
			if err != nil {
				return err
			}
		}

		// delete current files that exist in the chart
		fmt.Printf("Removing local files which are contained in the source chart...\n")

		for fi, err := range chartFileIter(sourceChartFiles.Iter()) {
			if err != nil {
				return err
			}
			if fi.Entry.IsDir() {
				continue
			}
			err = os.Remove(fi.FullPath)
			if err != nil {
				return err
			}
		}
	}

	// copy files from new chart
	fmt.Printf("Copying files from new version...\n")

	for fi, err := range chartFileIter(latestChartFiles.Iter()) {
		if err != nil {
			return err
		}
		if fi.Entry.IsDir() {
			continue
		}
		targetFile := filepath.Join(chartOutputPath, fi.Path)
		err = os.MkdirAll(filepath.Dir(targetFile), os.ModePerm)
		if err != nil {
			return err
		}

		err = file.CopyFile(fi.FullPath, targetFile)
		if err != nil {
			return err
		}
	}

	if !ignoreCurrent && applyPatch && !diffBuilder.IsEmpty() {
		// apply patch to new files
		patcher, err := diff.NewPatcher(diffBuilder.String())
		if err != nil {
			return fmt.Errorf("error loading patch file: %w", err)
		}

		for filediff := range patcher.Files() {
			targetFile := filepath.Join(chartOutputPath, filediff.NewName)
			targetFileData, err := os.ReadFile(targetFile)
			if os.IsNotExist(err) {
				fmt.Printf("patching %s failed: %s does not exist\n", filediff.NewName, targetFile)
				continue
			} else if err != nil {
				return err
			}

			// apply patch
			var output bytes.Buffer
			err = gitdiff.Apply(&output, bytes.NewReader(targetFileData), filediff)
			if err != nil {
				var fconflict *gitdiff.Conflict
				if errors.As(err, &fconflict) {
					fmt.Printf("conflict applying patch to %s: %s\n", filediff.NewName, err)

					conflictFileName, err := file.GenerateUniqueFilename(filepath.Dir(targetFile),
						file.NameExtFormat(filediff.NewName)+"_conflict", ".diff")
					if err != nil {
						return fmt.Errorf("error generating conflict file: %w", err)
					}

					if !file.Exists(conflictFileName) {
						err = os.WriteFile(conflictFileName, []byte(filediff.String()), os.ModePerm)
						if err != nil {
							return err
						}
					} else {
						fmt.Printf("could not write conflict patch to %s: file exists\n", conflictFileName)
					}

				} else {
					fmt.Printf("failed to apply patch to %s: %s\n", filediff.NewName, err)
				}
				continue
			}

			fmt.Printf("applied patch to %s\n", filediff.NewName)

			err = os.WriteFile(targetFile, output.Bytes(), os.ModePerm)
			if err != nil {
				return fmt.Errorf("error applying patch to %s: %w", filediff.NewName, err)
			}
		}
	}

	return nil
}
