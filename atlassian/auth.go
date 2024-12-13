package atlassian

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		client: httpClient,
	}
}

type Resource struct {
	ID        string   `json:"id"`        // "id": "1324a887-45db-1bf4-1e99-ef0ff456d421",
	Name      string   `json:"name"`      // "name": "Site name",
	URL       string   `json:"url"`       // "url": "https://your-domain.atlassian.net",
	Scopes    []string `json:"scopes"`    // "scopes": ["write:jira-work", "read:jira-user", "manage:jira-configuration"],
	AvatarUrl string   `json:"avatarUrl"` // "avatarUrl": "https:\/\/site-admin-avatar-cdn.prod.public.atl-paas.net\/avatars\/240\/flag.png"
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`

	response *http.Response
}

func (e *ErrorResponse) Error() string {
	base := "jira request failed: "
	if e.Message == "" || e.Code == 0 {
		return base + e.response.Status
	}
	return fmt.Sprintf("%s%d %s", base, e.Code, e.Message)
}

func (c *Client) requestJSON(method, url string, body io.Reader, v any) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		err = &ErrorResponse{
			response: resp,
		}
		jsonErr := json.Unmarshal(b, err)
		if jsonErr != nil {
			slog.Warn("failed to parse error json", "err", err)
		}
		return err
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AccessibleResources() ([]*Resource, error) {
	resources := []*Resource{}
	err := c.requestJSON(http.MethodGet, "https://api.atlassian.com/oauth/token/accessible-resources", http.NoBody, &resources)
	if err != nil {
		return nil, err
	}
	return resources, nil
}
