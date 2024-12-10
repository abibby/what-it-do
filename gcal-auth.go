package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abibby/what-it-do/ezoauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var GoogleConfig *ezoauth.Config

func init() {
	creds, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(creds, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}
	GoogleConfig = &ezoauth.Config{
		Name:        "google",
		OAuthConfig: config,
		AuthCodeURLOpts: []oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
		},
	}
}

func getGCalService(r *http.Request) (*calendar.Service, error) {
	client, err := GoogleConfig.Client(r)
	if err != nil {
		return nil, fmt.Errorf("could not start calendar client: %w", err)
	}

	srv, err := calendar.NewService(r.Context(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	return srv, nil
}
