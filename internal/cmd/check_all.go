package cmd

import (
	"context"
	"fmt"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/file"
	"github.com/rrgmc/helm-vendor/internal/helm"
	"helm.sh/helm/v3/pkg/repo"
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
	chartRoot, err := c.openChartRoot(chartConfig)
	if err != nil {
		return err
	}
	defer chartRoot.Close()

	// currentChartFilename := filepath.Join(c.buildChartPath(chartConfig), "Chart.yaml")
	currentChartFilename := "Chart.yaml"
	var currentChart *repo.ChartVersion
	if file.ExistsFS(chartRoot, currentChartFilename) {
		var err error
		currentChart, err = helm.LoadHelmChartVersionFile(chartRoot, currentChartFilename)
		if err != nil {
			return fmt.Errorf("error loading chart file %s: %w\n", currentChartFilename, err)
		}
	}

	repository, err := helm.LoadRepository(chartConfig.Repository.URL)
	if err != nil {
		return err
	}

	latestChart, err := repository.GetChart(chartConfig.Name, "")
	if err != nil {
		return err
	}

	fmt.Printf("- %s:", chartConfig.Path)
	if currentChart != nil {
		fmt.Printf(" [local:%s]", currentChart.Version)
	}
	fmt.Printf(" [latest:%s]", latestChart.Chart().Version)
	fmt.Printf("\n")

	return nil
}
