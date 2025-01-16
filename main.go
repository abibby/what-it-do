package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/abibby/salusa/clog"
)

const DateFormat = "January 2, 2006"

type Row struct {
	Date        time.Time
	Project     string
	SubCategory string
	Hours       time.Duration
	JiraID      string
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
		r.JiraID,
		r.Description,
	}
}

func main() {
	var levelInfo bool
	var levelDebug bool
	var cal bool
	var jira bool
	var bb bool
	var day string

	flag.BoolVar(&levelInfo, "v", false, "do verbose logging")
	flag.BoolVar(&levelDebug, "vv", false, "do verbose logging")
	flag.BoolVar(&cal, "calendar", false, "run calendar tests")
	flag.BoolVar(&jira, "jira", false, "run jira tests")
	flag.BoolVar(&bb, "bitbucket", false, "run bitbucket tests")
	flag.StringVar(&day, "date", time.Now().Format(time.DateOnly), "the date to get info for")

	flag.Parse()

	level := slog.LevelWarn
	if levelDebug {
		level = slog.LevelDebug
	} else if levelInfo {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(clog.DefaultHandler(level)))

	all := !cal && !jira && !bb

	now, err := time.Parse(time.DateOnly, day)
	check(err)
	start := startOfDay(now)
	end := endOfDay(now)

	out := csv.NewWriter(os.Stdout)
	out.Comma = '\t'

	rows := []*Row{}

	// if all || bb {
	// 	prRows, err := getCodeReviews(start, end)
	// 	check(err)
	// 	rows = append(rows, prRows...)
	// }
	if all || cal {
		calRows, err := addCalenderEvents(start, end)
		check(err)
		rows = append(rows, calRows...)
	}

	if all || jira {
		jiraRows, err := addJiraIssues(start, end)
		check(err)
		rows = append(rows, jiraRows...)
	}

	for _, row := range rows {
		err = out.Write(row.ToCSVRow())
		check(err)
	}
	err = out.Write([]string{})
	check(err)

	out.Flush()
}

func check(err error) {
	if err == nil {
		return
	}

	slog.Error("Fatal error", "err", err)
}
