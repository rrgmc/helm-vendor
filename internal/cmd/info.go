package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rrgmc/helm-vendor/internal/config"
	"github.com/rrgmc/helm-vendor/internal/file"
	"github.com/rrgmc/helm-vendor/internal/helm"
	"helm.sh/helm/v3/pkg/repo"
)

func (c *Cmd) Info(ctx context.Context, path string) error {
	for _, chartConfig := range c.cfg.Charts {
		if path == chartConfig.Path {
			return c.infoChart(ctx, chartConfig)
		}
	}
	return fmt.Errorf("unknown path '%s'", path)
}

func (c *Cmd) infoChart(ctx context.Context, chartConfig config.Chart) error {
	chartRoot, err := c.openChartRoot(chartConfig)
	if err != nil {
		return err
	}
	defer chartRoot.Close()

	fmt.Printf("%s:\n", chartConfig.Path)

	currentChartFilename := "Chart.yaml"
	var currentChart *repo.ChartVersion
	if file.Exists(chartRoot, currentChartFilename) {
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
	fmt.Printf("- description: %s\n", latestChart.Chart().Description)
	if currentChart != nil {
		fmt.Printf("- local: %s\n", currentChart.Version)
	}
	fmt.Printf("- latest: %s\n", helm.GetChartVersion(latestChart.Chart()))
	fmt.Printf("- versions:\n")
	for entry, err := range repository.ChartVersions(chartConfig.Name, 10) {
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
