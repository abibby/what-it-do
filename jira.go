package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/abibby/salusa/set"
	"github.com/abibby/what-it-do/config"
	"github.com/andygrunwald/go-jira"
)

const (
	FieldTestCases = "customfield_10034"
	FieldSprint    = "customfield_10020"
)

func addJiraIssues(start, end time.Time, out *csv.Writer) error {
	cfg := config.Load()

	tp := jira.BasicAuthTransport{
		Username: cfg.Jira.Username,
		Password: cfg.Jira.Password,
	}
	jiraClient, err := jira.NewClient(tp.Client(), cfg.Jira.BaseURL)
	if err != nil {
		return err
	}

	currentUser, _, err := jiraClient.User.GetSelf()
	if err != nil {
		return err
	}

	issues, _, err := jiraClient.Issue.Search(
		fmt.Sprintf("project = PD AND (assignee = currentUser() OR issuekey in updatedBy(\"%s\")) AND sprint in openSprints() ORDER BY created DESC", currentUser.DisplayName),
		&jira.SearchOptions{
			Fields: []string{"*all"},
		},
	)
	if err != nil {
		return err
	}

	date := time.Now()
	for _, issue := range issues {
		subCategory := ""

		changelogs, _, err := GetChangelogs(jiraClient, issue.ID, nil)
		if err != nil {
			return err
		}

		states := statesToday(&issue, changelogs.Values)

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
			Date:        date,
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

func statesToday(issue *jira.Issue, changes []*jira.ChangelogHistory) set.Set[string] {
	states := set.New[string]()
	after := startOfDay(time.Now()).Format(time.RFC3339)
	for _, change := range changes {
		if change.Created < after {
			continue
		}
		for _, item := range change.Items {
			if item.Field == "status" {
				states.Add(item.ToString)
			}
		}
	}
	states.Add(issue.Fields.Status.Name)
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

// type ChangelogItem struct {
// 	// The user who made the change.
// 	Author *jira.User `json:"author"`

// 	// The date on which the change took place.
// 	Created jira.Time `json:"created"`

//		HistoryMetadata *jira.ChangelogHistory `json:"historyMetadata"`
//	}
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
