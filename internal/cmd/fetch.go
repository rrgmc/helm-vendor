package cmd

import (
	"context"
	"fmt"
)

func (c *Cmd) Fetch(ctx context.Context, path string, version string) error {
	for _, chartConfig := range c.cfg.Charts {
		err := c.printVersionCheck(ctx, chartConfig)
		if err != nil {
			fmt.Printf("error checking chart: %s\n", err)
		}
	}
	return nil
}
