// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sendgrid provides the implementation of an email sender using SendGrid.
package sendgrid

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// Config for the SendGrid email provider.
type Config struct {
	APIKey      string `name:"api-key" description:"The SendGrid API key to use"`
	SandboxMode bool   `name:"sandbox" description:"Use SendGrid sandbox mode for testing"`
}

// SendGrid is the type that implements SendGrid as email provider.
type SendGrid struct {
	logger    log.Interface
	config    Config
	client    *sendgrid.Client
	fromEmail *mail.Email
}

// New creates a SendGrid email provider.
func New(logger log.Interface, emailConfig email.Config, sgConfig Config) (email.Sender, error) {
	provider := &SendGrid{
		logger:    logger.WithField("email_provider", "SendGrid"),
		config:    sgConfig,
		client:    sendgrid.NewSendClient(sgConfig.APIKey),
		fromEmail: mail.NewEmail(emailConfig.SenderName, emailConfig.SenderAddress),
	}
	return provider, nil
}

var errEmailNotSent = errors.DefineInternal("email_not_sent", "email was not sent")

// Send an email message.
func (s *SendGrid) Send(message *email.Message) error {
	logger := s.logger.WithFields(log.Fields(
		"template_name", message.TemplateName,
		"recipient_name", message.RecipientName,
		"recipient_address", message.RecipientAddress,
	))

	email, err := s.buildEmail(message)
	if err != nil {
		return err
	}

	logger.Debug("Sending email...")
	response, err := s.client.Send(email)
	if err != nil {
		return errEmailNotSent.WithCause(err)
	}

	if response.StatusCode >= 300 {
		attributes := []interface{}{
			"status_code", response.StatusCode,
			"response", response.Body,
		}
		logger.WithFields(log.Fields(attributes...)).WithError(err).Error("Could not send email")
		return errEmailNotSent.WithAttributes(attributes...)
	}

	logger.Info("Sent email")
	return nil
}

func (s *SendGrid) buildEmail(email *email.Message) (*mail.SGMailV3, error) {
	message := mail.NewV3MailInit(
		s.fromEmail,
		email.Subject,
		mail.NewEmail(email.RecipientName, email.RecipientAddress),
	)
	if email.TextBody != "" {
		message.AddContent(mail.NewContent("text/plain", email.TextBody))
	}
	if email.HTMLBody != "" {
		message.AddContent(mail.NewContent("text/html", email.HTMLBody))
	}
	if s.config.SandboxMode {
		settings := mail.NewMailSettings()
		settings.SetSandboxMode(mail.NewSetting(true))
		message = message.SetMailSettings(settings)
	}
	return message, nil
}
