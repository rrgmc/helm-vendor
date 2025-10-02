package helm

import (
	"os"
	"path/filepath"

	"github.com/rrgmc/helm-vendor/internal/file"
)

type ChartFiles struct {
	chart      *Chart
	path       string
	isTempPath bool
	chartRoot  *os.Root
}

func newChartFiles(chart *Chart, path string, isTempPath bool) (*ChartFiles, error) {
	chartRoot, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}
	return &ChartFiles{
		chart:      chart,
		path:       path,
		isTempPath: isTempPath,
		chartRoot:  chartRoot,
	}, nil
}

func (c *ChartFiles) Root() *os.Root {
	return c.chartRoot
}

func (c *ChartFiles) Iter() file.Iter {
	// return file.IterDir(c.chartRoot.FS(), filepath.Join(c.path, filepath.Clean(c.chart.chart.Name)))
	return file.IterDir(c.chartRoot.FS(), filepath.Clean(c.chart.chart.Name))
}

func (c *ChartFiles) Close() error {
	_ = c.chartRoot.Close()
	if c.isTempPath && c.path != "" {
		return os.RemoveAll(c.path)
	}
	return nil
}
