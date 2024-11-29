package main

import (
	"encoding/csv"
	"log"
	"os"
	"time"
)

const DateFormat = "January 2, 2006"

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
