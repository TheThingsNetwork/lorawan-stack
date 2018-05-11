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

package sendgrid

import (
	"github.com/jaytaylor/html2text"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/email/templates"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// Config is the type that configures the SendGrid email provider.
type Config struct {
	// APIKey is the sendgrid api key.
	APIKey string `name:"api-key" description:"The sendgrid API key to use"`

	// SandboxMode enables sendgrid's sandbox mode for testing.
	SandboxMode bool `name:"sandbox" description:"Set the sendgrid sandbox mode for testing"`

	// From is the address the emails are sent from.
	From string `name:"from" description:"The address the emails are sent from"`

	// Name is the name of the sender.
	Name string `name:"name" description:"The name of the sender"`
}

// SendGrid is the type that implements SendGrid as email provider.
type SendGrid struct {
	logger    log.Interface
	config    Config
	client    *sendgrid.Client
	fromEmail *mail.Email
}

// New creates a SendGrid email provider.
func New(logger log.Interface, config Config) *SendGrid {
	provider := &SendGrid{
		logger:    logger.WithField("provider", "SendGrid"),
		client:    sendgrid.NewSendClient(config.APIKey),
		fromEmail: mail.NewEmail(config.Name, config.From),
		config:    config,
	}

	return provider
}

// Send sends an email to recipient using the provided template along with the data.
func (s *SendGrid) Send(recipient string, template templates.Template) error {
	message, err := s.buildEmail(recipient, template)
	if err != nil {
		return err
	}

	logger := s.logger.WithFields(log.Fields(
		"recipient", recipient,
		"template_name", template.GetName(),
	))

	logger.Info("Sending email ...")

	response, err := s.client.Send(message)

	if err != nil {
		logger.WithFields(log.Fields(
			"status_code", response.StatusCode,
			"response", response.Body,
		)).WithError(err).Error("Failed to send email")

		return err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		logger.WithFields(log.Fields(
			"status_code", response.StatusCode,
			"response", response.Body,
		)).Error("Failed to send email")

		return errors.Errorf("Failed to send email. Status code `%d`", response.StatusCode)
	}

	logger.Info("Email successfully sent")

	return nil
}

// buildEmail builds the email that will be sent using the underlying SendGrid client.
func (s *SendGrid) buildEmail(recipient string, template templates.Template) (*mail.SGMailV3, error) {
	subject, content, err := template.Render()
	if err != nil {
		return nil, err
	}

	text, err := html2text.FromString(content, html2text.Options{PrettyTables: true})
	if err != nil {
		return nil, err
	}

	message := mail.NewV3MailInit(
		s.fromEmail,
		subject,
		mail.NewEmail("", recipient),
		mail.NewContent("text/html", content),
		mail.NewContent("text/plain", text),
	)

	if s.config.SandboxMode {
		settings := mail.NewMailSettings()
		settings.SetSandboxMode(mail.NewSetting(true))

		message = message.SetMailSettings(settings)
	}

	return message, nil
}
