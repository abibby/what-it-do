package ezoauth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/abibby/salusa/request"
	"golang.org/x/oauth2"
)

var ErrUnauthenticated = errors.New("Unauthenticated")

type Config struct {
	Name            string
	OAuthConfig     *oauth2.Config
	AuthCodeURLOpts []oauth2.AuthCodeOption
}

// Retrieve a token, saves the token, then returns the generated client.
func (c *Config) GetToken(r *http.Request) (*oauth2.Token, error) {
	cookies := r.CookiesNamed(c.cookieName())

	if len(cookies) < 1 {
		return nil, fmt.Errorf("no cookies found for %s: %w", c.Name, ErrUnauthenticated)
	}

	cookie := cookies[0]
	b, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.Unmarshal(b, tok)
	return tok, err
}

func (c *Config) Client(r *http.Request) (*http.Client, error) {
	tok, err := c.GetToken(r)
	if err != nil {
		return nil, err
	}
	return c.OAuthConfig.Client(r.Context(), tok), nil
}

func (c *Config) AuthCodeURL() string {
	return c.OAuthConfig.AuthCodeURL("state", c.AuthCodeURLOpts...)
}

// func newState() (string, error) {
// 	b := make([]byte, 32)
// 	_, err := rand.Reader.Read(b)
// 	if err != nil {
// 		return "", err
// 	}
// 	buf := bytes.Buffer{}
// 	_, err = base64.NewEncoder(base64.RawURLEncoding, &buf).Write(b)
// 	if err != nil {
// 		return "", err
// 	}
// 	return buf.String(), nil
// }

func (c *Config) Callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// if q.Get("state") != state {
	// 	w.WriteHeader(401)
	// 	return
	// }
	authCode := q.Get("code")
	tok, err := c.OAuthConfig.Exchange(r.Context(), authCode)
	if err != nil {
		request.RespondError(w, r, fmt.Errorf("unable to complete token exchange: %w", err))
		return
	}
	rawJSON, err := json.Marshal(tok)
	if err != nil {
		request.RespondError(w, r, fmt.Errorf("unable to add token cookie: %w", err))
		return
	}
	b64 := base64.URLEncoding.EncodeToString(rawJSON)

	u, err := url.Parse(c.OAuthConfig.RedirectURL)
	if err != nil {
		request.RespondError(w, r, fmt.Errorf("unable to find base domain: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   c.cookieName(),
		Value:  b64,
		Path:   "/",
		Domain: u.Hostname(),
	})
	// log.Printf("%s\n%s\n\n", c.cookieName(), b64)
	// http.Redirect(w, r, "/", http.StatusFound)
	fmt.Fprintf(w, "%d", len(b64))
}
func (c *Config) cookieName() string {
	return "oauth_" + c.Name
}
