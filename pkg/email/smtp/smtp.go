// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package smtp

import (
	"context"
	"crypto/tls"
	"net"
	"net/smtp"
	"strconv"

	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/log"
	gomail "gopkg.in/mail.v2"
)

// Config for the SMTP email provider.
type Config struct {
	Address     string `name:"address" description:"SMTP server address"`
	Username    string `name:"username" description:"Username to authenticate with"`
	Password    string `name:"password" description:"Password to authenticate with"`
	Connections int    `name:"connections" description:"Maximum number of connections to the SMTP server"`
	TLSConfig   *tls.Config
}

func (c Config) auth() smtp.Auth {
	if c.Username == "" && c.Password == "" {
		return nil
	}
	host, _, _ := net.SplitHostPort(c.Address)
	return smtp.PlainAuth("", c.Username, c.Password, host)
}

type sendTask struct {
	message *gomail.Message
	result  chan error
}

// SMTP is the type that implements SMTP as email provider.
type SMTP struct {
	ctx         context.Context
	logger      log.Interface
	emailConfig email.Config
	smtpConfig  Config
	dialer      *gomail.Dialer
	tasks       chan sendTask
}

func (s *SMTP) handle() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.tasks:
			task.result <- s.dialer.DialAndSend(task.message)
		}
	}
}

var buffer = 8 // send buffer per connection

// New creates a SMTP email provider.
func New(ctx context.Context, emailConfig email.Config, smtpConfig Config) (email.Sender, error) {
	host, portStr, _ := net.SplitHostPort(smtpConfig.Address)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	if smtpConfig.Connections == 0 {
		smtpConfig.Connections = 1
	}
	s := &SMTP{
		ctx:         ctx,
		logger:      log.FromContext(ctx).WithField("email_provider", "SMTP"),
		emailConfig: emailConfig,
		smtpConfig:  smtpConfig,
		dialer:      gomail.NewDialer(host, port, smtpConfig.Username, smtpConfig.Password),
		tasks:       make(chan sendTask, smtpConfig.Connections*buffer),
	}
	if smtpConfig.TLSConfig != nil {
		s.dialer.TLSConfig = smtpConfig.TLSConfig.Clone()
		s.dialer.TLSConfig.ServerName, _, _ = net.SplitHostPort(s.smtpConfig.Address)
	}
	for i := 0; i < smtpConfig.Connections; i++ {
		go s.handle()
	}
	return s, nil
}

// Send an email message.
func (s *SMTP) Send(message *email.Message) error {
	logger := s.logger.WithFields(log.Fields(
		"template_name", message.TemplateName,
		"recipient_name", message.RecipientName,
		"recipient_address", message.RecipientAddress,
	))

	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.emailConfig.SenderAddress, s.emailConfig.SenderName)
	m.SetAddressHeader("To", message.RecipientAddress, message.RecipientName)
	m.SetHeader("Subject", message.Subject)
	if message.TextBody != "" {
		m.AddAlternative("text/plain", message.TextBody)
	}
	if message.HTMLBody != "" {
		m.AddAlternative("text/html", message.HTMLBody)
	}

	sendResult := make(chan error)
	logger.Debug("Sending email...")
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case s.tasks <- sendTask{message: m, result: sendResult}:
		err := <-sendResult
		if err != nil {
			logger.WithError(err).Error("Could not send email")
			return err
		}
		logger.Info("Sent email")
		return nil
	}
}
