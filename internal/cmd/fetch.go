package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/file"
	"github.com/rrgmc/helm-vendor/internal/helm"
)

func (c *Cmd) Fetch(ctx context.Context, path string, version string) error {
	for _, chartConfig := range c.cfg.Charts {
		if path == chartConfig.Path {
			return c.fetchChart(ctx, chartConfig, version)
		}
	}
	return fmt.Errorf("unknown path '%s'", path)
}

func (c *Cmd) fetchChart(ctx context.Context, chartConfig config.Chart, version string) error {
	chartOutputPath := c.buildChartPath(chartConfig)
	currentChartFilename := filepath.Join(chartOutputPath, "Chart.yaml")
	if file.Exists(currentChartFilename) {
		return fmt.Errorf("chart already exists in path '%s', use upgrade to download a newer version", chartConfig.Path)
	}

	repo, err := helm.LoadRepository(chartConfig.Repository.URL)
	if err != nil {
		return err
	}

	chart, err := repo.GetChart(chartConfig.Name, version)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading '%s' [%s - %s]\n", chartConfig.Path, chart.Chart().Name, chart.Chart().Version)

	chartFiles, err := chart.Download()
	if err != nil {
		return err
	}
	defer chartFiles.Close()

	chartFileIter := func(iter file.Iter) file.Iter {
		return file.IterFilter(iter, file.Filter{
			Ignore: chartConfig.Files.Ignore,
		})
	}

	// copy files from chart
	for fi, err := range chartFileIter(chartFiles.Iter()) {
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

	return nil
}
