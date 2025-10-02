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
}

func (c *ChartFiles) Iter() file.Iter {
	return file.IterDir(filepath.Join(c.path, filepath.Clean(c.chart.chart.Name)))
}

func (c *ChartFiles) Close() error {
	if c.isTempPath && c.path != "" {
		return os.RemoveAll(c.path)
	}
	return nil
}
