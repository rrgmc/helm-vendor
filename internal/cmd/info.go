package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/helm"
	"helm.sh/helm/v3/pkg/repo"
)

func (c *Cmd) Info(ctx context.Context, path string, allVersions bool) error {
	for _, chartConfig := range c.cfg.Charts {
		if path == chartConfig.Path {
			return c.infoChart(ctx, chartConfig, allVersions)
		}
	}
	return fmt.Errorf("unknown path '%s'", path)
}

func (c *Cmd) infoChart(ctx context.Context, chartConfig config.Chart, allVersions bool) error {
	fmt.Printf("%s:\n", chartConfig.Path)

	currentChartFilename := "Chart.yaml"
	var currentChart *repo.ChartVersion
	if c.chartRootExists(chartConfig) {
		chartRoot, err := c.openChartRoot(chartConfig)
		if err != nil {
			return err
		}
		defer chartRoot.Close()

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
	if latestChart.Chart().Description != "" {
		fmt.Printf("- description: %s\n", latestChart.Chart().Description)
	}
	if currentChart != nil {
		fmt.Printf("- local: %s\n", currentChart.Version)
	} else {
		fmt.Printf("- local: not found\n")
	}
	fmt.Printf("- latest: %s\n", helm.GetChartVersion(latestChart.Chart()))
	fmt.Printf("- versions:\n")
	maxVersions := 10
	if allVersions {
		maxVersions = -1
	}
	for entry, err := range repository.ChartVersions(chartConfig.Name, maxVersions) {
		if err != nil {
			fmt.Printf("error listing chart versions: %s\n", err)
			break
		}
		var date string
		if !entry.Created.IsZero() {
			date = fmt.Sprintf(" [%s]", entry.Created.Format(time.RFC3339))
		}
		fmt.Printf("\t- %s%s\n", entry.Version, date)
	}

	return nil
}
