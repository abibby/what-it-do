package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/abibby/salusa/set"
	"github.com/abibby/what-it-do/atlassian"
	"github.com/abibby/what-it-do/ezoauth"
	"github.com/andygrunwald/go-jira"
	"golang.org/x/oauth2"
)

const (
	FieldTestCases = "customfield_10034"
	FieldSprint    = "customfield_10020"
)

type JiraOAuth struct {
	URL string
}

func ConfigFromJSON(b []byte) (*oauth2.Config, error) {
	type a struct {
		ClientID     string   `json:"client_id"`
		Scopes       []string `json:"scopes"`
		RedirectURI  string   `json:"redirect_uri"`
		ClientSecret string   `json:"client_secret"`
	}

	creds := &a{}
	err := json.Unmarshal(b, creds)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.atlassian.com/authorize",
			TokenURL: "https://auth.atlassian.com/oauth/token",
		},
		RedirectURL: creds.RedirectURI,
		Scopes:      creds.Scopes,
	}, nil
}

func getJiraClient() (*jira.Client, error) {
	ctx := context.Background()
	creds, err := os.ReadFile("atlassian_creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := ConfigFromJSON(creds)
	if err != nil {
		return nil, err
	}
	ezconfig := &ezoauth.Config{
		Name:        "jira",
		OAuthConfig: config,
		AuthCodeURLOpts: []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("audience", "api.atlassian.com"),
			oauth2.ApprovalForce,
		},
	}
	client, err := ezconfig.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start jira client: %w", err)
	}

	atlassianClient := atlassian.NewClient(client)
	resources, err := atlassianClient.AccessibleResources()
	if err != nil {
		return nil, err
	}

	if len(resources) != 1 {
		return nil, fmt.Errorf("more than one resource returned")
	}

	jiraClient, err := jira.NewClient(client, "https://api.atlassian.com/ex/jira/"+url.PathEscape(resources[0].ID))
	if err != nil {
		return nil, err
	}
	return jiraClient, nil
}

func addJiraIssues(start, end time.Time, out *csv.Writer) error {
	jiraClient, err := getJiraClient()
	if err != nil {
		return err
	}

	currentUser, _, err := GetMyself(jiraClient, nil)
	if err != nil {
		return fmt.Errorf("get self: %w", err)
	}

	issues, _, err := jiraClient.Issue.Search(
		fmt.Sprintf("project = PD AND (assignee = currentUser() OR issuekey in updatedBy(\"%s\")) AND sprint in openSprints() ORDER BY created DESC", currentUser.DisplayName),
		&jira.SearchOptions{
			Fields: []string{"*all"},
		},
	)
	if err != nil {
		return fmt.Errorf("issue search: %w", err)
	}

	for _, issue := range issues {
		subCategory := ""

		changelogs, _, err := GetChangelogs(jiraClient, issue.ID, nil)
		if err != nil {
			return fmt.Errorf("issue changelog: %w", err)
		}

		states := statesBetween(&issue, changelogs.Values, start, end)

		if issue.Fields.Assignee.AccountID == currentUser.AccountID {
			if states.Has("In Progress") {
				if issue.Fields.Type.Name == "Test Execution" {
					subCategory = "Testing"
				} else {
					subCategory = "Implementation"
				}
			}
		} else {
			if states.Has("In Testing") {
				if hasEditedField(changelogs.Values, currentUser.AccountID, "Test Cases") {
					subCategory = "Testing"
				}
			}
		}

		if subCategory == "" {
			continue
		}

		err = out.Write(Row{
			Date:        start,
			Project:     "Technical - ",
			SubCategory: subCategory,
			Description: fmt.Sprintf("%s: %s", issue.Key, issue.Fields.Summary),
		}.ToCSVRow())
		if err != nil {
			return err
		}
	}

	return nil
}

func hasEditedField(changes []*jira.ChangelogHistory, accountID string, field string) bool {
	for _, change := range changes {
		if change.Author.AccountID != accountID {
			continue
		}

		for _, item := range change.Items {
			if item.Field == field {
				return true
			}
		}
	}
	return false
}

func statesBetween(issue *jira.Issue, changes []*jira.ChangelogHistory, minTime, maxTime time.Time) set.Set[string] {
	states := set.New[string]()

	var lastBefore *jira.ChangelogHistory
	var firstAfter *jira.ChangelogHistory

	for _, change := range changes {
		created, err := change.CreatedTime()
		if err != nil {
			panic(err)
		}
		if created.Before(minTime) {
			lastBefore = change
			continue
		}
		if created.After(maxTime) {
			firstAfter = change
			break
		}
		for _, item := range change.Items {
			if item.Field == "status" {
				states.Add(item.FromString)
				states.Add(item.ToString)
			}
		}
	}

	if lastBefore != nil {
		for _, item := range lastBefore.Items {
			if item.Field == "status" {
				states.Add(item.ToString)
			}
		}
	}

	if firstAfter != nil {
		for _, item := range firstAfter.Items {
			if item.Field == "status" {
				states.Add(item.FromString)
			}
		}
	} else {
		states.Add(issue.Fields.Status.Name)
	}

	return states
}

type PaginatedResponse[T any] struct {
	// Whether this is the last page.
	IsLast bool `json:"isLast"`

	// The maximum number of items that could be returned.
	MaxResults int32 `json:"maxResults"`

	// If there is another page of results, the URL of the next page.
	NextPage string `json:"nextPage"`

	// The index of the first item returned.
	Self string `json:"self"`

	// The index of the first item returned.
	StartAt int64 `json:"startAt"`

	// The number of items returned.
	Total int32 `json:"total"`

	// The list of items.
	Values []T `json:"values"`
}

type MyselfOptions struct {
	Expand string
}

func GetMyself(client *jira.Client, options *MyselfOptions) (*jira.User, *jira.Response, error) {
	return GetMyselfContext(context.Background(), client, options)
}
func GetMyselfContext(ctx context.Context, client *jira.Client, options *MyselfOptions) (*jira.User, *jira.Response, error) {
	u := url.URL{
		Path: "/rest/api/3/myself",
	}
	uv := url.Values{}

	if options != nil {
		if options.Expand != "" {
			uv.Add("expand", options.Expand)
		}
	}

	u.RawQuery = uv.Encode()

	req, err := client.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	v := &jira.User{}
	resp, err := client.Do(req, v)
	if err != nil {
		err = jira.NewJiraError(resp, err)
	}
	return v, resp, err
}

type ChangelogResponse PaginatedResponse[*jira.ChangelogHistory]
type ChangelogOptions struct {
	StartAt    int
	MaxResults int
}

func GetChangelogs(client *jira.Client, id string, options *ChangelogOptions) (*ChangelogResponse, *jira.Response, error) {
	return GetChangelogsContext(context.Background(), client, id, options)
}
func GetChangelogsContext(ctx context.Context, client *jira.Client, id string, options *ChangelogOptions) (*ChangelogResponse, *jira.Response, error) {
	u := url.URL{
		Path: fmt.Sprintf("/rest/api/3/issue/%s/changelog", url.PathEscape(id)),
	}
	uv := url.Values{}

	if options != nil {
		if options.StartAt != 0 {
			uv.Add("startAt", strconv.Itoa(options.StartAt))
		}
		if options.MaxResults != 0 {
			uv.Add("maxResults", strconv.Itoa(options.MaxResults))
		}
	}

	u.RawQuery = uv.Encode()

	req, err := client.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	v := &ChangelogResponse{}
	resp, err := client.Do(req, v)
	if err != nil {
		err = jira.NewJiraError(resp, err)
	}
	return v, resp, err
}
