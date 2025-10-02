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
	chartRoot, err := c.openChartRoot(chartConfig)
	if err != nil {
		return err
	}
	defer chartRoot.Close()

	currentChartFilename := "Chart.yaml"
	if file.Exists(chartRoot, currentChartFilename) {
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
		err = chartRoot.MkdirAll(filepath.Dir(fi.Path), os.ModePerm)
		if err != nil {
			return err
		}

		err = file.CopyFile(chartFiles.Root(), chartRoot, fi.Path, fi.Path)
		if err != nil {
			return err
		}
	}

	return nil
}
