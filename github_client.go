package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/go-github/v50/github"
)

type GithubClient struct {
	*github.Client
}

func (gh *GithubClient) DownloadHelmChartsFromOrgs(ctx context.Context, orgs string) error {
	repos, err := gh.GetRepositories(ctx, orgs)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		_, err := os.Stat("./out/" + repo.GetName())
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			continue
		}

		if err := gh.DownloadDir(ctx, repo, "chart", "./out/"+repo.GetName()); err != nil {
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
