package app

import (
	"context"

	"github.com/abibby/salusa/event"
	"github.com/abibby/salusa/event/cron"
	"github.com/abibby/salusa/kernel"
	"github.com/abibby/salusa/openapidoc"
	"github.com/abibby/salusa/salusadi"
	"github.com/abibby/salusa/view"
	"github.com/abibby/what-it-do/app/events"
	"github.com/abibby/what-it-do/app/jobs"
	"github.com/abibby/what-it-do/app/models"
	"github.com/abibby/what-it-do/app/providers"
	"github.com/abibby/what-it-do/config"
	"github.com/abibby/what-it-do/migrations"
	"github.com/abibby/what-it-do/resources"
	"github.com/abibby/what-it-do/routes"
	"github.com/abibby/what-it-do/services/sms"
	"github.com/go-openapi/spec"
	"github.com/google/uuid"
)

var Kernel = kernel.New(
	kernel.Config(config.Load),
	kernel.Bootstrap(
		salusadi.Register[*models.User](migrations.Use()),
		sms.Register,
		view.Register(resources.Content, "**/*.html"),
		providers.Register,
		func(ctx context.Context) error {
			openapidoc.RegisterFormat[uuid.UUID]("uuid")
			return nil
		},
	),
	kernel.Services(
		cron.Service().
			Schedule("* * * * *", &events.LogEvent{Message: "cron event"}),
		event.Service(
			event.NewListener[*jobs.LogJob](),
		),
	),
	kernel.InitRoutes(routes.InitRoutes),
	kernel.APIDocumentation(
		openapidoc.Info(spec.InfoProps{
			Title:       "Salusa Example API",
			Description: `This is the API documentaion for the example Salusa application`,
		}),
		openapidoc.BasePath("/api"),
	),
)
