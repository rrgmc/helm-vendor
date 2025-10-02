package cmd

import (
	"context"
	"fmt"
)

func (c *Cmd) FetchAll(ctx context.Context) error {
	for _, chartConfig := range c.cfg.Charts {
		if c.chartRootFileExists(chartConfig) {
			continue
		}
		err := c.fetchChart(ctx, chartConfig, "")
		if err != nil {
			fmt.Printf("error fetching chart: %s\n", err)
		}
	}
	return nil
}
