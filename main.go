package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
			Usage:   "get helm charts",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "gh-access-token",
					Aliases: []string{"gh-token", "t"},
					Usage:   "Github Access Token",
					EnvVars: []string{"GH_ACCESS_TOKEN"},
				},
			},
			Action: func(ctx *cli.Context) error {
				ts := oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: ctx.String("gh-access-token")},
				)
				tc := oauth2.NewClient(ctx.Context, ts)

				client := &GithubClient{github.NewClient(tc)}

				err := fetchCharts(context.Background(), client)
				if err != nil {
					return err
				}
				return nil
			},
		},
		&cli.Command{
			Name:  "verify",
			Usage: "Verify charts",
			Action: func(ctx *cli.Context) error {
				files, err := os.ReadDir("./out")
				if err != nil {
					log.Fatal(err)
				}
				for _, file := range files {
					out, err := render("./out/"+file.Name()+"/chart", "values-staging.yaml")
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
	cmd := exec.Command("conftest", "test", "--no-fail", chart)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func fetchCharts(ctx context.Context, client *GithubClient) error {
	repos, err := client.GetRepositories(ctx, "SadaPay")
	if err != nil {
		return err
	}

	for _, repo := range repos {
		// if !strings.Contains(*repo.Name, "") {
		// 	continue
		// }

		_, err := os.Stat("./out/" + repo.GetName())
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			continue
		}

		if err := client.DownloadDir(ctx, repo, "chart", "./out/"+repo.GetName()); err != nil {
			os.RemoveAll("./out/" + repo.GetName())
			fmt.Println(err)
			continue
		}
	}

	return nil
}

func (gh *GithubClient) DownloadDir(ctx context.Context, repo *github.Repository, dir string, outputDir string) error {
	os.MkdirAll(outputDir+"/"+dir, 0755)

	_, dirContent, _, err := gh.Repositories.GetContents(ctx, repo.Owner.GetLogin(), repo.GetName(), dir, &github.RepositoryContentGetOptions{})
	if err != nil {
		return err
	}
	for _, content := range dirContent {
		if content.DownloadURL == nil {
			if err := gh.DownloadDir(ctx, repo, dir+"/"+content.GetName(), outputDir); err != nil {
				return err
			}
			continue
		}

		rc, _, err := gh.Repositories.DownloadContents(ctx, repo.Owner.GetLogin(), repo.GetName(), content.GetPath(), nil)
		if err != nil {
			return err
		}

		fmt.Println(outputDir + "/" + content.GetPath())
		f, err := os.Create(outputDir + "/" + content.GetPath())
		if err != nil {
			return err
		}
		_, err = io.Copy(f, rc)
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

type GithubClient struct {
	*github.Client
}

func (gh *GithubClient) GetRepositories(ctx context.Context, orgName string) ([]*github.Repository, error) {
	var result []*github.Repository

	f, err := os.Open("repositories.json")
	if err == nil {
		err = json.NewDecoder(f).Decode(&result)
		return result, err
	}
	f.Close()

	for page := 1; ; page++ {
		opt := &github.RepositoryListByOrgOptions{
			Sort: "name",
			ListOptions: github.ListOptions{
				PerPage: 100,
				Page:    page,
			},
		}
		repos, _, err := gh.Repositories.ListByOrg(ctx, orgName, opt)
		if err != nil {
			return nil, err
		}
		result = append(result, repos...)
		if len(repos) == 0 {
			break
		}
	}

	f, err = os.Create("repositories.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func render(path string, valueFile string) ([]byte, error) {
	cmd := exec.Command("helm", "template", "-f", path+"/"+valueFile, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, err
}
