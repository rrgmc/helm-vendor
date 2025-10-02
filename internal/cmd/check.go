package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/file"
	"github.com/rrgmc/helm-vendor/internal/helm"
)

func (c *Cmd) Check(ctx context.Context, path string) error {
	for _, chartConfig := range c.cfg.Charts {
		if path == chartConfig.Path {
			return c.checkChart(ctx, chartConfig)
		}
	}
	return fmt.Errorf("unknown path '%s'", path)
}

func (c *Cmd) checkChart(ctx context.Context, chartConfig config.Chart) error {
	fmt.Printf("%s:\n", chartConfig.Path)

	currentChartFilename := filepath.Join(c.buildChartPath(chartConfig), "Chart.yaml")
	if file.Exists(currentChartFilename) {
		currentChart, err := helm.LoadHelmChartVersionFile(currentChartFilename)
		if err != nil {
			return fmt.Errorf("error loading chart file %s: %w\n", currentChartFilename, err)
		}
		fmt.Printf("- local version: %s\n", currentChart.Version)
	}

	repo, err := helm.LoadRepository(chartConfig.Repository.URL)
	if err != nil {
		return err
	}

	latestChart, err := repo.GetChart(chartConfig.Name, "")
	if err != nil {
		return err
	}
	fmt.Printf("- description: %s\n", latestChart.Chart().Description)
	fmt.Printf("- latest version: %s\n", latestChart.Chart().Version)
	fmt.Printf("- versions:\n")
	for entry, err := range repo.ChartVersions(chartConfig.Name, 10) {
		if err != nil {
			return err
		}
		fmt.Printf("\t- %s [%s]\n", entry.Version, entry.Created.Format(time.DateOnly))
	}

	return nil
}
