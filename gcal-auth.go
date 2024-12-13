package main

import (
	"context"
	"fmt"
	"os"

	"github.com/abibby/what-it-do/config"
	"github.com/abibby/what-it-do/ezoauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func getGCalService() (*calendar.Service, error) {
	ctx := context.Background()
	creds, err := os.ReadFile(config.Dir("google_creds.json"))
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(creds, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	ezconfig := &ezoauth.Config{
		Name:        "google",
		OAuthConfig: config,
		AuthCodeURLOpts: []oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
		},
	}
	client, err := ezconfig.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start calendar client: %w", err)
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	return srv, nil
}
