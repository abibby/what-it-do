package jobs

import (
	"context"
	"log/slog"

	"github.com/abibby/what-it-do/app/events"
)

type LogJob struct {
	Logger *slog.Logger `inject:""`
}

func (l *LogJob) Handle(ctx context.Context, e *events.LogEvent) error {
	l.Logger.Info(e.Message)
	return nil
}
