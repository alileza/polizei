package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

var Get = &cli.Command{
	Name:    "get",
	Aliases: []string{"g"},
	Usage:   "Get helm charts from github orgs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "token",
			Aliases:  []string{"t"},
			Required: true,
			Usage:    "GitHub access token",
			EnvVars:  []string{"GH_ACCESS_TOKEN"},
		},
		&cli.StringFlag{
			Name:     "org",
			Aliases:  []string{"o"},
			Required: true,
			Usage:    "GitHub organization name",
			EnvVars:  []string{"GH_ORG"},
		},
		&cli.StringFlag{
			Name:    "pattern",
			Aliases: []string{"p"},
			Value:   "",
			Usage:   "Regex pattern to match repository names",
			EnvVars: []string{"GH_REPO_PATTERN"},
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"out"},
			Value:   "./out",
			Usage:   "Output directory for cloned repositories",
		},
	},
	Action: func(c *cli.Context) error {
		org := c.String("org")
		pattern := c.String("pattern")
		output := c.String("output")
		token := c.String("token")

		// Retrieve the list of repositories for the given organization
		reposOutput, err := getAllRepos(token, org)
		if err != nil {
			return err
		}
		// Iterate over the list of repositories and clone them if they don't already exist in the output directory
		repos := string(reposOutput)
		reposSlice := regexp.MustCompile(`\r?\n`).Split(repos, -1)
		for _, repo := range reposSlice {
			if repo == "" {
				continue
			}

			if pattern != "" {
				match, _ := regexp.MatchString(pattern, repo)
				if !match {
					continue
				}
			}

			repoName := repo[strings.LastIndex(repo, "/")+1 : len(repo)-4]
			repoPath := fmt.Sprintf("%s/%s", output, repoName)

			// Check if the repository already exists in the output directory
			if _, err := os.Stat(repoPath); err == nil {
				fmt.Printf("%s already exists in %s. Skipping...\n", repoName, output)
				continue
			}

			fmt.Printf("Cloning %s...\n", repo)
			cloneCmd := exec.Command("git", "clone", "--recursive", repo, repoPath)
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			err = cloneCmd.Run()
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func getAllRepos(token string, org string) (string, error) {
	var reposOutput []byte

	// Retrieve the first page of repositories for the given organization
	reposCmd := exec.Command("sh", "-c", fmt.Sprintf("curl -s -H 'Authorization: token %s' 'https://api.github.com/orgs/%s/repos?per_page=100&page=1' | jq -r '.[].clone_url'", token, org))
	reposOutputPage, err := reposCmd.Output()
	if err != nil {
		return "", err
	}

	// Append the first page of results to the output
	reposOutput = append(reposOutput, reposOutputPage...)

	// Retrieve the next page of repositories until there are no more pages
	page := 2
	for {
		reposCmd = exec.Command("sh", "-c", fmt.Sprintf("curl -s -H 'Authorization: token %s' 'https://api.github.com/orgs/%s/repos?per_page=100&page=%d' | jq -r '.[].clone_url'", token, org, page))
		nextReposOutput, err := reposCmd.Output()
		if err != nil {
			return "", err
		}

		if len(nextReposOutput) == 0 {
			// We've reached the end of the pages, so break out of the loop
			break
		}

		// Append the next page of results to the output
		reposOutput = append(reposOutput, nextReposOutput...)
		page++
	}

	return string(reposOutput), nil
}
