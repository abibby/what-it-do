package bitbucket

import "time"

type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type AccountLinks struct {
	Avatar *Link `json:"avatar"`
}

type Account struct {
	Type        string        `json:"type"`
	Links       *AccountLinks `json:"links"`
	CreatedOn   string        `json:"created_on"`
	DisplayName string        `json:"display_name"`
	UUID        string        `json:"uuid"`
}

type Participant struct {
	User           *Account `json:"user"`
	Role           string   `json:"role"`
	Approved       bool     `json:"approved"`
	State          string   `json:"state"`
	ParticipatedOn string   `json:"participated_on"`
}

type PullRequest struct {
	// Links *PullRequestLinks `json:"links"`
	ID    int    `json:"id"`
	Title string `json:"title"`
	// Rendered    *RenderedPullRequestMarkup `json:"rendered"`
	Summary any      `json:"summary"`
	State   string   `json:"state"`
	Author  *Account `json:"author"`
	// Source      *PullRequestEndpoint       `json:"source"`
	// Destination *PullRequestEndpoint       `json:"destination"`
	// MergeCommit *PullRequestCommit         `json:"merge_commit"`

	// // https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-get // links object
	// id integer
	// title string
	// rendered Rendered Pull Request Markup
	// summary object
	// state string
	// author allOf [object, Account]
	// source Pull Request Endpoint
	// destination Pull Request Endpoint
	// merge_commit Pull Request Commit
	// comment_count integer
	// task_count integer
	// close_source_branch boolean
	// closed_by allOf [object, Account]
	// reason string
	// created_on string
	// updated_on string

	// The list of users that were added as reviewers on this pull request when
	// it was created. For performance reasons, the API only includes this list
	// on a pull request's self URL.
	Reviewers []*Account `json:"reviewers"`

	// The list of users that are collaborating on this pull request.
	// Collaborators are user that:
	//
	// * are added to the pull request as a reviewer (part of the reviewers
	//   list)
	// * are not explicit reviewers, but have commented on the pull request
	// * are not explicit reviewers, but have approved the pull request
	//
	// Each user is wrapped in an object that indicates the user's role and
	// whether they have approved the pull request. For performance reasons,
	// the API only returns this list when an API requests a pull request by
	// id.
	Participants []*Participant `json:"participants"`
}

type PullRequestActivity struct {
	PullRequest *PullRequest         `json:"pull_request"`
	Approval    *PullRequestApproval `json:"approval"`
	Comment     *PullRequestComment  `json:"comment"`
}

type PullRequestApproval struct {
	Date time.Time `json:"date"`
}
type PullRequestComment struct {
	Date time.Time `json:"date"`
}

type RepositoryLinks struct {
	Self         *Link   `json:"self"`
	Html         *Link   `json:"html"`
	Avatar       *Link   `json:"avatar"`
	Pullrequests *Link   `json:"pullrequests"`
	Commits      *Link   `json:"commits"`
	Forks        *Link   `json:"forks"`
	Watchers     *Link   `json:"watchers"`
	Downloads    *Link   `json:"downloads"`
	Clone        []*Link `json:"clone"`
	Hooks        *Link   `json:"hooks"`
}

type Repository struct {
	Links       *RepositoryLinks `json:"links"`
	UUID        string           `json:"uuid"`
	FullName    string           `json:"full_name"`
	IsPrivate   bool             `json:"is_private"`
	Parent      *Repository      `json:"parent"`
	Scm         string           `json:"scm"`
	Owner       *Account         `json:"owner"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	CreatedOn   string           `json:"created_on"`
	UpdatedOn   string           `json:"updated_on"`

	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/#api-repositories-workspace-get
	// links object
	// uuid string
	// full_name string
	// is_private boolean
	// parent allOf [object, Repository]
	// scm string
	// owner allOf [object, Account]
	// name string
	// description string
	// created_on string
	// updated_on string
	// size integer
	// language string
	// has_issues boolean
	// has_wiki boolean
	// fork_policy string
	// project allOf [object, Project]
	// mainbranch allOf [Ref, Branch]
}
