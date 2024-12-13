package ezoauth

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

type Creds struct {
	ClientID     string   `json:"client_id"`
	Scopes       []string `json:"scopes"`
	RedirectURI  string   `json:"redirect_uri"`
	ClientSecret string   `json:"client_secret"`
}

func ReadConfigJSON(path string) (*oauth2.Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %w", err)
	}

	creds := &Creds{}
	err = json.Unmarshal(b, creds)
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
