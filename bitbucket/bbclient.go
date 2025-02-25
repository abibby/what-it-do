package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"

	"github.com/abibby/what-it-do/jsonio"
)

type Client struct {
	httpClient      *http.Client
	baseURL         string
	internalBaseURL string
}

func NewClient(c *http.Client) *Client {
	return &Client{
		httpClient:      c,
		baseURL:         "https://api.bitbucket.org",
		internalBaseURL: "https://bitbucket.org",
	}
}

func (c *Client) CurrentUser() (*Account, error) {
	u := &Account{}
	err := c.request(http.MethodGet, "/2.0/user", nil, nil, u)
	return u, err
}

type ListPullRequestsOptions struct {
	Workspace string
	Slug      string
	Fields    string   `query:"fields"`
	State     []string `query:"state"`
	Query     string   `query:"q"`
}

func (c *Client) ListPullRequests(options *ListPullRequestsOptions) (*PaginatedResponse[*PullRequest], error) {
	u := &PaginatedResponse[*PullRequest]{
		client: c,
	}
	err := c.request(http.MethodGet, "/2.0/repositories/"+url.PathEscape(options.Workspace)+"/"+url.PathEscape(options.Slug)+"/pullrequests", options, nil, u)
	return u, err
}

func (c *Client) ListPullRequestActivity(options *ListPullRequestsOptions) (*PaginatedResponse[*PullRequestActivity], error) {
	u := &PaginatedResponse[*PullRequestActivity]{
		client: c,
	}
	err := c.request(http.MethodGet, "/2.0/repositories/"+url.PathEscape(options.Workspace)+"/"+url.PathEscape(options.Slug)+"/pullrequests/activity", options, nil, u)
	return u, err
}

type ListRepositoriesOptions struct {
	Workspace string
	Fields    string `query:"fields"`
	Role      string `query:"role"`
	Query     string `query:"q"`
	Sort      string `query:"sort"`
}

func (c *Client) ListRepositories(options *ListRepositoriesOptions) (*PaginatedResponse[*Repository], error) {
	u := &PaginatedResponse[*Repository]{
		client: c,
	}
	err := c.request(http.MethodGet, "/2.0/repositories/"+url.PathEscape(options.Workspace), options, nil, u)
	return u, err
}

func (c *Client) request(method, p string, query, body any, v any) error {
	var bodyReader io.Reader = http.NoBody
	if body != nil {
		bodyReader = jsonio.NewReader(body)
	}

	queryValues := toValues(query)

	return c.rawRequest(method, c.baseURL+p+"?"+queryValues.Encode(), bodyReader, v)
}

func (c *Client) internalRequest(method, p string, query, body any, v any) error {
	var bodyReader io.Reader = http.NoBody
	if body != nil {
		bodyReader = jsonio.NewReader(body)
	}

	queryValues := toValues(query)

	return c.rawRequest(method, c.internalBaseURL+p+"?"+queryValues.Encode(), bodyReader, v)
}

func (c *Client) rawRequest(method, url string, body io.Reader, v any) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			b = []byte("unknown error")
		}
		return fmt.Errorf("fetch error: %s", b)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
func toValues(v any) url.Values {
	rv := reflect.ValueOf(v)
	if (rv == reflect.Value{} || rv.IsNil()) {
		return url.Values{}
	}

	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	if rt.Kind() != reflect.Struct {
		panic("v must be a struct or a pointer to a struct")
	}

	val := url.Values{}
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		q := sf.Tag.Get("query")
		if q == "" {
			continue
		}

		fv := rv.Field(i)

		if (fv == reflect.Value{} || (fv.IsZero())) {
			continue
		}
		if rv.Kind() == reflect.Array {
			for j := 0; j < rv.Len(); j++ {
				val.Add(q, fmt.Sprint(fv.Index(i).Interface()))
			}
		} else {
			val.Add(q, fmt.Sprint(fv.Interface()))
		}
	}

	return val
}

type ListWorkspacePullRequestsOptions struct {
	Workspace string
	Fields    string `query:"fields"`
	Query     string `query:"q"`
}

func (c *Client) ListWorkspacePullRequests(options *ListWorkspacePullRequestsOptions) (*PaginatedResponse[*PullRequest], error) {
	u := &PaginatedResponse[*PullRequest]{
		client: c,
	}
	err := c.internalRequest(http.MethodGet, "/!api/internal/workspaces/"+url.PathEscape(options.Workspace)+"/pullrequests/", options, nil, u)
	return u, err
}

// https://bitbucket.org/!api/internal/workspaces/ownersbox/pullrequests/?fields=-values.closed_by%2C-values.description%2C-values.summary%2C-values.rendered%2C-values.properties%2C-values.reason%2C-values.reviewers%2C-values.participants.user.nickname%2C%2Bvalues.destination.branch.name%2C%2Bvalues.destination.repository.full_name%2C%2Bvalues.destination.repository.name%2C%2Bvalues.destination.repository.uuid%2C%2Bvalues.destination.repository.full_name%2C%2Bvalues.destination.repository.name%2C%2Bvalues.destination.repository.links.self.href%2C%2Bvalues.destination.repository.links.html.href%2C%2Bvalues.source.branch.name%2C%2Bvalues.source.repository.full_name%2C%2Bvalues.source.repository.name%2C%2Bvalues.source.repository.uuid%2C%2Bvalues.source.repository.full_name%2C%2Bvalues.source.repository.name%2C%2Bvalues.source.repository.links.self.href%2C%2Bvalues.source.repository.links.html.href%2C%2Bvalues.source.commit.hash&page=1&pagelen=20&q=state%3D%22OPEN%22%20AND%20followers.uuid%3D%22c8c46d4c-e199-4859-a510-e038ef88d80e%22
