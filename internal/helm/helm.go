package helm

import (
	"io"
	"os"

	"github.com/rrgmc/helm-vendor/internal/yaml"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func LoadHelmChartVersion(r io.Reader) (*repo.ChartVersion, error) {
	var chart repo.ChartVersion
	if err := yaml.Decode(r, &chart); err != nil {
		return nil, err
	}
	return &chart, nil
}

func LoadHelmChartVersionFilename(filename string) (*repo.ChartVersion, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadHelmChartVersion(f)
}

func LoadHelmChartVersionFile(root *os.Root, filename string) (*repo.ChartVersion, error) {
	var chart repo.ChartVersion
	if err := yaml.DecodeFile(root, filename, &chart); err != nil {
		return nil, err
	}
	return &chart, nil
}

func GetChartVersion(chartVersion *repo.ChartVersion) string {
	ver := chartVersion.Version
	if ver == "" {
		ver = "unknown"
	}
	return ver
}

var allGetters = getter.All(&cli.EnvSettings{})
