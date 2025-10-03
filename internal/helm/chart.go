package helm

import (
	"errors"
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/repo"
)

type Chart struct {
	repository *Repository
	chart      *repo.ChartVersion
}

func LoadChart(repository *Repository, chart *repo.ChartVersion) (*Chart, error) {
	return &Chart{
		repository: repository,
		chart:      chart,
	}, nil
}

func (c *Chart) Chart() *repo.ChartVersion {
	return c.chart
}

func (c *Chart) Download(options ...ChartDownloadOption) (*ChartFiles, error) {
	var optns chartDownloadOptions
	for _, opt := range options {
		opt(&optns)
	}

	var chartURL string

	if len(c.chart.URLs) == 0 {
		return nil, errors.New("chart has no downloadable URLs")
	}

	chartURL = c.chart.URLs[0]

	absoluteChartURL, err := c.repository.ResolveReferenceURL(chartURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make chart URL absolute: %w", err)
	}

	isTempPath := false
	if optns.downloadPath == "" {
		optns.downloadPath, err = os.MkdirTemp("", "helm-chart")
		if err != nil {
			return nil, fmt.Errorf("unable to create temporary directory for download: %w", err)
		}
		isTempPath = true
	}

	dl := downloader.ChartDownloader{
		Out:            os.Stderr,
		Getters:        allGetters,
		RegistryClient: c.repository.registry,
	}

	chartPackageFile, _, err := dl.DownloadTo(absoluteChartURL, c.chart.Version, optns.downloadPath)
	if err != nil {
		return nil, fmt.Errorf("error downloading chart: %w", err)
	}

	err = chartutil.ExpandFile(optns.downloadPath, chartPackageFile)
	if err != nil {
		return nil, fmt.Errorf("error expanding chart: %w", err)
	}

	_ = os.Remove(chartPackageFile)
	// if err != nil {
	// 	return nil, fmt.Errorf("error removing chart temporary file: %w", err)
	// }

	return newChartFiles(c, optns.downloadPath, isTempPath)
}

func WithChartDownloadPath(path string) ChartDownloadOption {
	return func(options *chartDownloadOptions) {
		options.downloadPath = path
	}
}

type ChartDownloadOption func(*chartDownloadOptions)

type chartDownloadOptions struct {
	downloadPath string
}
