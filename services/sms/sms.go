package sms

import (
	"context"
	"fmt"

	"github.com/abibby/salusa/di"
	"github.com/abibby/salusa/salusaconfig"
)

type SMSConfiger interface {
	SMSConfig() Config
}

type Config interface {
	Client() Client
}

type Client interface {
	Send(to, msg string) error
}

func Register(ctx context.Context) error {
	di.RegisterLazySingletonWith(ctx, func(cfg salusaconfig.Config) (Client, error) {
		cfger, ok := any(cfg).(SMSConfiger)
		if !ok {
			return nil, fmt.Errorf("config not instance of sms.SMSConfiger")
		}
		return cfger.SMSConfig().Client(), nil
	})
	return nil
}
