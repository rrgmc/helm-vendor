package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rrgmc/helm-vendor/internal/helm"
)

func Download(ctx context.Context, repoURL string, name string, version string, allVersions bool,
	outputValuesFile bool, outputPath string) error {
	repo, err := helm.LoadRepository(repoURL)
	if err != nil {
		return err
	}

	latestChart, err := repo.GetChart(name, version)
	if err != nil {
		return err
	}

	var descPrefix string
	if outputValuesFile {
		descPrefix = "# helm-vendor: "
	}

	fmt.Printf("%s%s:\n", descPrefix, latestChart.Chart().Name)

	if latestChart.Chart().Description != "" {
		fmt.Printf("%s- description: %s\n", descPrefix, latestChart.Chart().Description)
	}
	if version == "" {
		fmt.Printf("%s- latest: %s\n", descPrefix, helm.GetChartVersion(latestChart.Chart()))
	} else {
		fmt.Printf("%s- requested version: %s\n", descPrefix, helm.GetChartVersion(latestChart.Chart()))
	}
	if !outputValuesFile {
		fmt.Printf("- versions:\n")
		maxVersions := 15
		if allVersions {
			maxVersions = -1
		}
		for entry, err := range repo.ChartVersions(name, maxVersions) {
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
	}
	if outputPath != "" {
		fmt.Printf("%sWriting chart files to %s...\n", descPrefix, outputPath)
		latestChartFiles, err := latestChart.Download(helm.WithChartDownloadPath(outputPath))
		if err != nil {
			return err
		}
		defer latestChartFiles.Close()
	}

	if outputValuesFile {
		latestChartFiles, err := latestChart.Download()
		if err != nil {
			return err
		}
		defer latestChartFiles.Close()

		for fi, err := range latestChartFiles.Iter() {
			if err != nil {
				return err
			}
			if fi.Entry.IsDir() {
				continue
			}
			if fi.Path == "values.yaml" {
				fd, err := latestChartFiles.Root().ReadFile(fi.Path)
				if err != nil {
					return err
				}
				fmt.Println(string(fd))
				return nil
			}
		}
		return fmt.Errorf("no values.yaml found in this chart")
	}

	return nil
}
