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
	end := startOfDay(now.Add(24 * time.Hour))

	out := csv.NewWriter(os.Stdout)
	out.Comma = '\t'

	if all || cal {
		err := addCalenderEvents(start, end, out)
		if err != nil {
			log.Fatal(err)
		}
	}

	if all || jira {
		err := addJiraIssues(start, end, out)
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
