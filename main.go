package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
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

func main() {
	var cal bool
	var jira bool
	var day string

	flag.BoolVar(&cal, "calendar", false, "run calendar tests")
	flag.BoolVar(&jira, "jira", false, "run jira tests")
	flag.StringVar(&day, "date", time.Now().Format(time.DateOnly), "the date to get info for")

	flag.Parse()

	all := !cal && !jira

	now, err := time.Parse(time.DateOnly, day)
	if err != nil {
		log.Fatal(err)
	}
	start := startOfDay(now)
	end := endOfDay(now)

	out := csv.NewWriter(os.Stdout)
	out.Comma = '\t'

	rows := []*Row{}
	if all || cal {
		calRows, err := addCalenderEvents(start, end)
		if err != nil {
			log.Fatal(err)
		}
		rows = append(rows, calRows...)
	}

	if all || jira {
		jiraRows, err := addJiraIssues(start, end)
		if err != nil {
			log.Fatal(err)
		}
		rows = append(rows, jiraRows...)
	}

	for _, row := range rows {
		err = out.Write(row.ToCSVRow())
		if err != nil {
			log.Fatal(err)
		}
	}
	err = out.Write([]string{})
	if err != nil {
		log.Fatal(err)
	}

	out.Flush()
}
