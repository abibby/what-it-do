package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abibby/what-it-do/bitbucket"
	"github.com/abibby/what-it-do/config"
	"github.com/abibby/what-it-do/ezoauth"
	"golang.org/x/oauth2"
)

var jiraRE = regexp.MustCompile(`PD-\d+`)

func getCodeReviews(start, end time.Time) ([]*Row, error) {
	bbCient, err := getBitbucketService()
	if err != nil {
		return nil, err
	}
	u, err := bbCient.CurrentUser()
	if err != nil {
		return nil, err
	}

	rows := []*Row{}

	prs, err := bbCient.ListWorkspacePullRequests(&bitbucket.ListWorkspacePullRequestsOptions{
		Workspace: "ownersbox",
		Fields:    "+reviewers",
		Query:     fmt.Sprintf(`(state="MERGED" or state="OPEN") and followers.uuid="%s" and updated_on > %s AND updated_on < %s`, u.UUID, start.Format(time.RFC3339), end.Format(time.RFC3339)),
	})
	if err != nil {
		return nil, err
	}

	for pr := range prs.All() {

		if !didReview(pr, u, start, end) {
			continue
		}

		description := pr.Title
		jiraID := jiraRE.FindString(description)
		if jiraID != "" {
			description = regexp.MustCompile(".*"+jiraID+":?").ReplaceAllString(description, "")
		}
		description = strings.TrimSpace(description)
		rows = append(rows, &Row{
			Date:        start,
			Project:     "Technical - ",
			SubCategory: "Code Review",
			JiraID:      jiraID,
			Description: description,
		})
	}

	return rows, nil
}

func didReview(pr *bitbucket.PullRequest, user *bitbucket.Account, start, end time.Time) bool {
	for _, par := range pr.Participants {
		if par.User.UUID != user.UUID {
			continue
		}
		participatedOn, err := time.Parse(time.RFC3339, par.ParticipatedOn)
		if err != nil {
			return false
		}
		if start.Before(participatedOn) && end.After(participatedOn) {
			return true
		}
	}
	return false
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
