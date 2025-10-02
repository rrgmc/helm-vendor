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

func LoadHelmChartVersionFile(fsys *os.Root, filename string) (*repo.ChartVersion, error) {
	var chart repo.ChartVersion
	if err := yaml.DecodeFile(fsys, filename, &chart); err != nil {
		return nil, err
	}
	return &chart, nil
}

var allGetters = getter.All(&cli.EnvSettings{})
