package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rrgmc/helm-vendor/internal/helm"
)

func Download(ctx context.Context, repoURL string, name string, version string, outputPath string) error {
	repo, err := helm.LoadRepository(repoURL)
	if err != nil {
		return err
	}

	latestChart, err := repo.GetChart(name, version)
	if err != nil {
		return err
	}

	fmt.Printf("%s:\n", latestChart.Chart().Name)

	if latestChart.Chart().Description != "" {
		fmt.Printf("- description: %s\n", latestChart.Chart().Description)
	}
	fmt.Printf("- latest: %s\n", helm.GetChartVersion(latestChart.Chart()))
	fmt.Printf("- versions:\n")
	for entry, err := range repo.ChartVersions(name, 15) {
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

	if outputPath != "" {
		fmt.Printf("Writing chart files to %s...\n", outputPath)
		latestChartFiles, err := latestChart.Download(helm.WithChartDownloadPath(outputPath))
		if err != nil {
			return err
		}
		defer latestChartFiles.Close()
	}

	return nil
}
