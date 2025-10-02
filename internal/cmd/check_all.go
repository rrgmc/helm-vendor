package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/file"
	"github.com/rrgmc/helm-vendor/internal/helm"
)

func (c *Cmd) CheckAll(ctx context.Context) error {
	for _, chartConfig := range c.cfg.Charts {
		err := c.runCheckAll(ctx, chartConfig)
		if err != nil {
			fmt.Printf("error checking chart: %s\n", err)
		}
	}
	return nil
}

func (c *Cmd) runCheckAll(ctx context.Context, chartConfig config.Chart) error {
	fmt.Printf("- %s:", chartConfig.Path)

	currentChartFilename := filepath.Join(c.buildChartPath(chartConfig), "Chart.yaml")
	if file.Exists(currentChartFilename) {
		currentChart, err := helm.LoadHelmChartVersionFile(currentChartFilename)
		if err != nil {
			return fmt.Errorf("error loading chart file %s: %w\n", currentChartFilename, err)
		}
		fmt.Printf(" [local:%s]", currentChart.Version)
	}

	repo, err := helm.LoadRepository(chartConfig.Repository.URL)
	if err != nil {
		return err
	}

	latestChart, err := repo.GetChart(chartConfig.Name, "")
	if err != nil {
		return err
	}
	fmt.Printf(" [latest:%s]", latestChart.Chart().Version)

	fmt.Printf("\n")

	return nil
}
