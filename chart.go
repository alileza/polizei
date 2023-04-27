package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
)

var Chart = &cli.Command{
	Name:        "chart",
	Aliases:     []string{"c"},
	Usage:       "List repository names that contain a 'chart' directory in their root",
	Subcommands: []*cli.Command{ChartRender},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"out"},
			Value:   "./out",
			Usage:   "Output directory for cloned repositories",
		},
		&cli.BoolFlag{
			Name:    "interactive",
			Aliases: []string{"i"},
			Usage:   "Interactive mode",
		},
	},
	Action: func(c *cli.Context) error {
		output := c.String("output")
		repos, err := ioutil.ReadDir(output)
		if err != nil {
			return err
		}

		var repoNames []string
		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			chartPath := fmt.Sprintf("%s/%s/chart", output, repo.Name())
			_, err := os.Stat(chartPath)
			if err == nil {
				repoNames = append(repoNames, repo.Name())
			}
		}

		if len(repoNames) == 0 {
			fmt.Println("No repositories found with a 'chart' directory")
			return nil
		}
		if !c.Bool("interactive") {
			for _, v := range repoNames {
				fmt.Println(v)
			}
			return nil
		}
		var chartName string
		prompt := &survey.Select{
			Message: "Select a repository to render:",
			Options: repoNames,
			Default: repoNames[0],
		}
		if err := survey.AskOne(prompt, &chartName, survey.WithValidator(survey.Required)); err != nil {
			return err
		}

		return nil
	},
}
