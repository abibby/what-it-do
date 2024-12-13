package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abibby/what-it-do/bitbucket"
	"github.com/abibby/what-it-do/config"
	"github.com/abibby/what-it-do/ezoauth"
	"golang.org/x/oauth2"
)

func getCodeReviews(start, end time.Time) ([]*Row, error) {
	bbCient, err := getBitbucketService()
	if err != nil {
		return nil, err
	}
	u, err := bbCient.CurrentUser()
	if err != nil {
		return nil, err
	}

	repos, err := bbCient.ListRepositories(&bitbucket.ListRepositoriesOptions{
		Workspace: "ownersbox",
		Role:      "contributor",
		Query:     fmt.Sprintf(`updated_on > %s AND updated_on < %s`, start.Format(time.RFC3339), end.Format(time.RFC3339)),
		Sort:      "-updated_on",
	})
	if err != nil {
		return nil, err
	}

	rows := []*Row{}
	for repo := range repos.All() {
		fmt.Println(repo.Name)
		repoPRs, err := bbCient.ListPullRequests(&bitbucket.ListPullRequestsOptions{
			Workspace: "ownersbox",
			Slug:      repo.UUID,
			Fields:    "+reviewers",
			State:     []string{"OPEN", "MERGED"},
			Query:     fmt.Sprintf(`reviewers.uuid="%s" and updated_on > %s AND updated_on < %s`, u.UUID, start.Format(time.RFC3339), end.Format(time.RFC3339)),
		})
		if err != nil {
			return nil, err
		}

		for pr := range repoPRs.All() {
			slog.Info("Reviewers", "len", len(pr.Reviewers))
			// for _, par := range pr.Participants {
			// }

			rows = append(rows, &Row{
				Date:        start,
				Project:     "Technical - ",
				SubCategory: "Code Review",
				Description: pr.Title,
			})
		}
	}

	return rows, nil
}

func getBitbucketService() (*bitbucket.Client, error) {
	ctx := context.Background()
	config, err := ezoauth.ReadConfigJSON(config.Dir("bitbucket_creds.json"))
	if err != nil {
		return nil, err
	}
	config.Endpoint = oauth2.Endpoint{
		AuthURL:  "https://bitbucket.org/site/oauth2/authorize",
		TokenURL: "https://bitbucket.org/site/oauth2/access_token",
	}
	ezconfig := &ezoauth.Config{
		Name:        "bitbucket",
		OAuthConfig: config,
	}
	client, err := ezconfig.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start bitbucket client: %w", err)
	}

	return bitbucket.NewClient(client), nil
}
