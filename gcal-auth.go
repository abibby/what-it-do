package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "embed"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

//go:embed credentials.json
var creds []byte

func getGCalService() (*calendar.Service, error) {
	ctx := context.Background()
	// b, err := os.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(creds, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client, err := getClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("could not start calendar client: %w", err)
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	return srv, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to get token from web: %w", err)
		}
		err = saveToken(tokFile, tok)
		if err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprintf(os.Stderr, "Go to the following link in your browser: \n%v\n", authURL)

	authCode, err := runCodePullServer(config)
	if err != nil {
		return nil, fmt.Errorf("code retrieval server failed: %w", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Fprintf(os.Stderr, "Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}
	return nil
}

func runCodePullServer(config *oauth2.Config) (string, error) {
	redirectURL, err := url.Parse(config.RedirectURL)
	if err != nil {
		log.Fatal(err)
	}

	authCode := ""
	var s *http.Server
	s = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !r.URL.Query().Has("code") {
				return
			}
			authCode = r.URL.Query().Get("code")

			fmt.Fprintf(w, "Continue in the terminal")

			go func() {
				time.Sleep(20 * time.Millisecond)
				s.Close()
			}()
		}),
		Addr: ":" + redirectURL.Port(),
	}

	err = s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		// no-op
	} else if err != nil {
		return "", err
	}
	return authCode, nil
}
