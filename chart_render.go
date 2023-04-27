package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var ChartRender = &cli.Command{
	Name:  "render",
	Usage: "Render a Helm chart using 'helm template'",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "directory",
			Aliases: []string{"d"},
			Value:   "./out",
			Usage:   "Directory containing cloned repositories",
		},
		&cli.StringFlag{
			Name:    "environment",
			Aliases: []string{"e"},
			Value:   "production",
			Usage:   "Environment to use for values file",
		},
		&cli.StringFlag{
			Name:    "output-file",
			Aliases: []string{"o"},
			Value:   "",
			Usage:   "File path to save rendered template, default to stdout",
		},
	},
	Action: func(c *cli.Context) error {
		directory := c.String("directory")
		env := c.String("environment")
		chart := c.Args().First()

		chartPath := fmt.Sprintf("%s/%s/chart", directory, chart)
		if _, err := os.Stat(chartPath); os.IsNotExist(err) {
			return fmt.Errorf("chart directory does not exist for repository '%s'", chart)
		}

		valuesPath := fmt.Sprintf("%s/%s/chart/values.yaml", directory, chart)
		_, err := os.Stat(valuesPath)
		if err != nil {
			fmt.Printf("Warning: values.yaml file not found for chart '%s'\n", chart)
		}

		valuesEnvPath := fmt.Sprintf("%s/%s/chart/values-%s.yaml", directory, chart, env)
		_, err = os.Stat(valuesEnvPath)
		if err != nil {
			fmt.Printf("Warning: values-%s.yaml file not found for chart '%s'\n", env, chart)
		}

		templateCmd := exec.Command("helm", "template", chartPath)
		if _, err = os.Stat(valuesEnvPath); err == nil {
			templateCmd.Args = append(templateCmd.Args, "-f", valuesEnvPath)
		} else if _, err = os.Stat(valuesPath); err == nil {
			templateCmd.Args = append(templateCmd.Args, "-f", valuesPath)
		}
		templateOutput, err := templateCmd.Output()
		if err != nil {
			fmt.Printf("Path: %s\nError: %s\nOutput: %s\n", chartPath, err.Error(), string(templateOutput))
			return err
		}

		outputFile := c.String("output-file")
		if outputFile != "" {
			if err := ioutil.WriteFile(outputFile, templateOutput, 0644); err != nil {
				return err
			}
		} else {
			fmt.Fprint(os.Stdout, string(templateOutput))
		}
		return nil
	},
}
