package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetCalenderEvents(r *http.Request, start, end time.Time) ([]*Row, error) {
	calendarService, err := getGCalService(r)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("unable to retrieve next ten of the user's events: %w", err)
	}

	rows := []*Row{}
	for _, item := range events.Items {
		if item.Start.Date != "" || item.End.Date != "" {
			continue
		}
		start, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for start: %w", err)
		}
		end, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for end: %w", err)
		}

		project := "Meetings - "
		description := item.Summary

		if strings.Contains(item.Summary, "Standup") {
			project = "Meetings - Daily Standup"
			description = ""
		} else if strings.Contains(item.Summary, "Sprint Demo") {
			project = "Meetings - Sprint Demo"
			description = ""
		} else if strings.Contains(item.Summary, "Backlog Refinement") {
			project = "Meetings - Backlog Refinement"
			description = ""
		}

		rows = append(rows, &Row{
			Date:        start,
			Project:     project,
			Hours:       end.Sub(start),
			Description: description,
		})
	}

	return rows, nil
}
