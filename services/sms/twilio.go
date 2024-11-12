package sms

import (
	"encoding/json"
	"fmt"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioConfig struct {
	AccountSid string
	AuthToken  string
	From       string
}

var _ Config = (*TwilioConfig)(nil)

// Client implements Config.
func (c *TwilioConfig) Client() Client {
	return &TwilioClient{
		c: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: c.AccountSid,
			Password: c.AuthToken,
		}),
		from: c.From,
	}
}

type TwilioClient struct {
	c    *twilio.RestClient
	from string
}

var _ Client = (*TwilioClient)(nil)

// Send implements Client.
func (c *TwilioClient) Send(to, msg string) error {

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(c.from)
	params.SetBody(msg)

	resp, err := c.c.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("error sending SMS message: %w", err)
	}

	response, _ := json.Marshal(*resp)
	fmt.Println("Response: " + string(response))
	return nil
}
