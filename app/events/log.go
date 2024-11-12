package events

import (
	"github.com/abibby/salusa/event"
	"github.com/abibby/salusa/event/cron"
)

type LogEvent struct {
	cron.CronEvent
	Message string
}

var _ event.Event = (*LogEvent)(nil)

func (e *LogEvent) Type() event.EventType {
	return "template:example-event"
}
