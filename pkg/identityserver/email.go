// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package identityserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/email/dir"
	"go.thethings.network/lorawan-stack/v3/pkg/email/sendgrid"
	"go.thethings.network/lorawan-stack/v3/pkg/email/smtp"
	_ "go.thethings.network/lorawan-stack/v3/pkg/email/templates" // Register all email templates.
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/sync/errgroup"
)

// SendEmail sends an email.
func (is *IdentityServer) SendEmail(ctx context.Context, message *email.Message) (err error) {
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"to", message.RecipientAddress,
		"subject", message.Subject,
		"template_name", message.TemplateName,
		"body", message.TextBody,
	))
	isConfig := is.configFromContext(ctx)
	var sender email.Sender
	switch isConfig.Email.Provider {
	case "sendgrid":
		sender, err = sendgrid.New(ctx, isConfig.Email.Config, isConfig.Email.SendGrid)
	case "smtp":
		sender, err = smtp.New(ctx, isConfig.Email.Config, isConfig.Email.SMTP)
	case "dir":
		sender, err = dir.New(ctx, isConfig.Email.Config, isConfig.Email.Dir)
	}
	if err != nil {
		logger.WithError(err).Warn("Could not send email without email provider")
		return err
	}
	if sender == nil {
		logger.Warn("Could not send email without email provider")
		return nil
	}
	err = sender.Send(message)
	if err != nil {
		logger.WithError(err).Warn("Failed to send email")
		return err
	}
	return nil
}

// SendTemplateEmailToUsers sends an email to users.
func (is *IdentityServer) SendTemplateEmailToUsers(
	ctx context.Context,
	templateName ttnpb.NotificationType,
	dataBuilder email.TemplateDataBuilder,
	receivers ...*ttnpb.User,
) error {
	networkConfig := is.configFromContext(ctx).Email.Network
	emailTemplate := email.GetTemplate(ctx, templateName)

	var wg errgroup.Group
	for _, receiver := range receivers {
		receiver := receiver // shadow range variable.
		wg.Go(func() error {
			templateData, err := dataBuilder(
				ctx,
				email.NewTemplateData(&networkConfig, receiver),
			)
			if err != nil {
				return err
			}
			message, err := emailTemplate.Execute(templateData)
			if err != nil {
				return err
			}
			return is.SendEmail(ctx, message)
		})
	}
	return wg.Wait()
}

// SendNotificationEmailToUsers sends a notification email to users.
func (is *IdentityServer) SendNotificationEmailToUsers(ctx context.Context, notification *ttnpb.Notification, receivers ...*ttnpb.User) error {
	networkConfig := is.configFromContext(ctx).Email.Network
	emailNotification := email.GetNotification(ctx, notification.GetNotificationType())
	emailTemplate := email.GetTemplate(ctx, emailNotification.EmailTemplateName)

	var wg errgroup.Group
	for _, receiver := range receivers {
		receiver := receiver // shadow range variable.
		wg.Go(func() error {
			templateData, err := emailNotification.DataBuilder(
				ctx,
				email.NewNotificationTemplateData(email.NewTemplateData(&networkConfig, receiver), notification),
			)
			if err != nil {
				return err
			}
			message, err := emailTemplate.Execute(templateData)
			if err != nil {
				return err
			}
			return is.SendEmail(ctx, message)
		})
	}
	return wg.Wait()
}

var emailUserFields = store.FieldMask{"ids", "name", "primary_email_address"}

// SendTemplateEmailToUserIDs looks up the users and sends them an email.
func (is *IdentityServer) SendTemplateEmailToUserIDs(
	ctx context.Context,
	templateName ttnpb.NotificationType,
	dataBuilder email.TemplateDataBuilder,
	receiverIDs ...*ttnpb.UserIdentifiers,
) error {
	var receivers []*ttnpb.User
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		receivers, err = st.FindUsers(ctx, receiverIDs, emailUserFields)
		return err
	})
	if err != nil {
		return err
	}
	return is.SendTemplateEmailToUsers(ctx, templateName, dataBuilder, receivers...)
}

var notificationEmailUserFields = store.FieldMask{"ids", "name", "primary_email_address", "admin"}

// SendNotificationEmailToUserIDs looks up the users and sends them a notification email.
func (is *IdentityServer) SendNotificationEmailToUserIDs(ctx context.Context, notification *ttnpb.Notification, receiverIDs ...*ttnpb.UserIdentifiers) error {
	var receivers []*ttnpb.User
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		receivers, err = st.FindUsers(ctx, receiverIDs, notificationEmailUserFields)
		return err
	})
	if err != nil {
		return err
	}
	return is.SendNotificationEmailToUsers(ctx, notification, receivers...)
}
