package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/abibby/what-it-do/ezoauth"
)

const DateFormat = "January 2, 2006"

type Row struct {
	Date        time.Time
	Project     string
	SubCategory string
	Hours       time.Duration
	Description string
}

func (r Row) ToCSVRow() []string {
	hours := ""
	if r.Hours != 0 {
		hours = fmt.Sprint(r.Hours.Hours())
	}
	return []string{
		r.Date.Format(DateFormat),
		r.Project,
		r.SubCategory,
		hours,
		r.Description,
	}
}

func rows(config *ezoauth.Config, handle func(r *http.Request, start, end time.Time) ([]*Row, error)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		start := startOfDay(now)
		end := endOfDay(now)
		rows, err := handle(r, start, end)
		if errors.Is(err, ezoauth.ErrUnauthenticated) {
			http.Redirect(w, r, config.AuthCodeURL(), http.StatusFound)
			return
		} else if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "failed to load jira issues: %v", err)
			return
		}

		err = json.NewEncoder(w).Encode(rows)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "failed to load jira issues: %v", err)
			return
		}
	})
}

func main() {
	http.HandleFunc("/issues", rows(JiraConfig, GetJiraIssues))
	http.HandleFunc("/events", rows(GoogleConfig, GetCalenderEvents))
	http.HandleFunc("/google/callback", GoogleConfig.Callback)
	http.HandleFunc("/atlassian/callback", JiraConfig.Callback)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, c := range r.Cookies() {
			fmt.Fprintf(w, "%s: %s\n\n", c.Name, c.Value)
			if strings.HasPrefix(c.Name, "oauth") {
			}
		}
	})
	http.ListenAndServe(":48663", nil)
}
