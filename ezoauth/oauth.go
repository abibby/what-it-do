package ezoauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/abibby/what-it-do/config"
	"golang.org/x/oauth2"
)

type Config struct {
	Name            string
	OAuthConfig     *oauth2.Config
	AuthCodeURLOpts []oauth2.AuthCodeOption
	ExchangeOpts    []oauth2.AuthCodeOption
}

type LogRoundTripper struct {
	Service   string
	Transport http.RoundTripper
}

var _ http.RoundTripper = (*LogRoundTripper)(nil)

// RoundTrip implements http.RoundTripper.
func (l *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	slog.Debug("Authenticated request", "service", l.Service, "url", req.URL, "method", req.Method)
	return l.Transport.RoundTrip(req)
}

func (c *Config) Client(ctx context.Context) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := path.Join(config.Dir(c.Name + "_token.json"))
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = c.getTokenFromWeb(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get token from web: %w", err)
		}
		err = saveToken(tokFile, tok)
		if err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	if !tok.Valid() {
		slog.Info("Refreshing oauth token", "service", c.Name)

		refreshed, err := c.OAuthConfig.TokenSource(ctx, tok).Token()
		if err != nil {
			slog.Warn("Failed to refresh access token", "err", err)
			tok, err = c.getTokenFromWeb(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get token from web: %w", err)
			}
		}

		if !tokensEqual(tok, refreshed) {
			err = saveToken(tokFile, tok)
			if err != nil {
				return nil, fmt.Errorf("failed to save token: %w", err)
			}
		}
		tok = refreshed
	}
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tok))

	client.Transport = &LogRoundTripper{Transport: client.Transport, Service: c.Name}
	return client, nil
}

func newState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Reader.Read(b)
	if err != nil {
		return "", err
	}
	buf := bytes.Buffer{}
	_, err = base64.NewEncoder(base64.RawURLEncoding, &buf).Write(b)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Request a token from the web, then returns the retrieved token.
func (c *Config) getTokenFromWeb(ctx context.Context) (*oauth2.Token, error) {
	state, err := newState()
	if err != nil {
		return nil, err
	}

	authURL := c.OAuthConfig.AuthCodeURL(state, c.AuthCodeURLOpts...)
	fmt.Fprintf(os.Stderr, "Go to the following link in your browser: \n%v\n", authURL)

	authCode, err := runCodePullServer(c.OAuthConfig, state)
	if err != nil {
		return nil, fmt.Errorf("code retrieval server failed: %w", err)
	}

	tok, err := c.OAuthConfig.Exchange(ctx, authCode, c.ExchangeOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to complete token exchange: %w", err)
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
func saveToken(p string, token *oauth2.Token) error {
	slog.Info("Saving credential file to", "file", p)

	err := os.MkdirAll(path.Dir(p), 0755)
	if err != nil {
		return fmt.Errorf("unable to create directory for oauth token: %v", err)
	}

	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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

func runCodePullServer(config *oauth2.Config, state string) (string, error) {
	redirectURL, err := url.Parse(config.RedirectURL)
	if err != nil {
		log.Fatal(err)
	}

	authCode := ""
	var s *http.Server
	s = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if !q.Has("code") {
				return
			}
			if q.Get("state") != state {
				w.WriteHeader(401)
				return
			}
			authCode = q.Get("code")

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

func tokensEqual(a, b *oauth2.Token) bool {
	return a.AccessToken == b.AccessToken &&
		a.RefreshToken == b.RefreshToken &&
		a.ExpiresIn == b.ExpiresIn &&
		a.Expiry == b.Expiry &&
		a.TokenType == b.TokenType
}
