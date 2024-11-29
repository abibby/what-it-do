package main

import (
	"encoding/csv"
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

	now := time.Now()
	start := startOfDay(now)
	end := startOfDay(now.Add(24 * time.Hour))

	out := csv.NewWriter(os.Stdout)
	out.Comma = '\t'

	err := addCalenderEvents(start, end, out)
	if err != nil {
		log.Fatal(err)
	}
	err = addJiraIssues(start, end, out)
	if err != nil {
		log.Fatal(err)
	}

	err = out.Write([]string{})
	if err != nil {
		log.Fatal(err)
	}

	out.Flush()
}
