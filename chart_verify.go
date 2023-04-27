package main

import (
	"fmt"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var ChartVerify = &cli.Command{
	Name:  "verify",
	Usage: "Verify a chart against a policy",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "chart-file",
			Aliases: []string{"f"},
			Usage:   "Chart file to verify",
		},
	},
	Action: func(c *cli.Context) error {
		chartFile := c.String("chart-file")

		if chartFile == "" {
			return cli.Exit("Chart file path is required", 1)
		}

		conftestCmd := exec.Command("conftest", "test", chartFile)
		conftestOutput, err := conftestCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Conftest output:\n%s\n", string(conftestOutput))
			return cli.Exit(fmt.Sprintf("Failed to verify chart: %v", err), 1)
		}

		fmt.Printf("Chart '%s' verified successfully!\n", chartFile)
		return nil
	},
}
