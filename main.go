package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/go-github/v50/github"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		&cli.Command{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "Get helm charts from github orgs",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "access-token",
					Aliases:     []string{"token", "t"},
					Usage:       "Github Access Token",
					DefaultText: "Retrieve and obtain helm charts, which are package managers for Kubernetes, from specific Github organizations. This process involves searching through the specified Github organization's repositories and locating the charts that are available for use in a Kubernetes cluster. Once the charts have been identified, they can be easily downloaded and implemented within the cluster for deployment and management of the various applications and services.",
					Required:    true,
					EnvVars:     []string{"GH_ACCESS_TOKEN"},
				},
				&cli.StringFlag{
					Name:     "orgs",
					Aliases:  []string{"o"},
					Usage:    "Github Organization",
					Required: true,
					EnvVars:  []string{"GH_ORGS"},
				},
			},
			Action: func(ctx *cli.Context) error {
				client := &GithubClient{
					github.NewClient(
						oauth2.NewClient(ctx.Context, oauth2.StaticTokenSource(
							&oauth2.Token{AccessToken: ctx.String("gh-access-token")},
						)),
					),
				}

				err := client.DownloadHelmChartsFromOrgs(ctx.Context, ctx.String("orgs"))
				if err != nil {
					return err
				}
				return nil
			},
		},
		&cli.Command{
			Name:    "verify",
			Usage:   "Verify charts",
			Aliases: []string{"v"},
			Action: func(ctx *cli.Context) error {
				files, err := os.ReadDir("./out")
				if err != nil {
					log.Fatal(err)
				}
				for _, file := range files {
					out, err := render("./out/"+file.Name()+"/chart", "./out/"+file.Name()+"/chart/values-staging.yaml")
					if err != nil {
						return fmt.Errorf("render: %w", err)
					}

					if err := os.WriteFile(fmt.Sprintf("chart/%s.yaml", file.Name()), out, 0755); err != nil {
						return fmt.Errorf("writeFile: %w", err)
					}

					chartName := fmt.Sprintf("chart/%s.yaml", file.Name())
					res, err := testChart(ctx.Context, chartName)
					if err != nil {
						return fmt.Errorf("testChart: %w", err)
					}
					fmt.Println(res)
				}
				return nil
			},
		},
	}
	app.RunAndExitOnError()
}

func testChart(ctx context.Context, chart string) (string, error) {
	cmd := exec.Command("conftest", "test", chart)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run()
	return out.String(), nil
}

func render(path string, valueFile string) ([]byte, error) {
	cmd := exec.Command("helm", "template", "-f", valueFile, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, err
}
