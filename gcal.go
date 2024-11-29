package main

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

func addCalenderEvents(start, end time.Time, out *csv.Writer) error {
	calendarService, err := getGCalService()
	if err != nil {
		return err
	}
	events, err := calendarService.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		MaxResults(100).
		OrderBy("startTime").
		Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve next ten of the user's events: %w", err)
	}

	for _, item := range events.Items {
		if item.Start.Date != "" || item.End.Date != "" {
			continue
		}
		start, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			return fmt.Errorf("invalid date format for start: %w", err)
		}
		end, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			return fmt.Errorf("invalid date format for end: %w", err)
		}

		project := "Meetings - "
		description := item.Summary

		if strings.Contains(item.Summary, "Standup") {
			project = "Meetings - Daily Standup"
			description = ""
		}

		err = out.Write(Row{
			Date:        start,
			Project:     project,
			Hours:       end.Sub(start),
			Description: description,
		}.ToCSVRow())
		if err != nil {
			return err
		}
	}

	return nil
}
