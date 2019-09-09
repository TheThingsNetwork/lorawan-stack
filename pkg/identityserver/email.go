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

package identityserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/email/sendgrid"
	"go.thethings.network/lorawan-stack/pkg/email/smtp"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (is *IdentityServer) getEmailTemplates(ctx context.Context) (*email.TemplateRegistry, error) {
	c := &is.configFromContext(ctx).Email.Templates
	c.registryMu.Lock()
	defer c.registryMu.Unlock()
	if c.registry != nil {
		return c.registry, nil
	}
	var fetcher fetch.Interface
	switch {
	case c.Static != nil:
		fetcher = fetch.NewMemFetcher(c.Static)
	case c.Directory != "":
		fetcher = fetch.FromFilesystem(c.Directory)
	case c.URL != "":
		var err error
		fetcher, err = fetch.FromHTTP(c.URL, true)
		if err != nil {
			return nil, err
		}
	default:
		fetcher = nil
	}
	c.registry = email.NewTemplateRegistry(fetcher, c.Includes...)
	return c.registry, nil
}

// SendEmail sends an email.
func (is *IdentityServer) SendEmail(ctx context.Context, f func(emails.Data) email.MessageData) (err error) {
	isConfig := is.configFromContext(ctx)
	var sender email.Sender
	switch isConfig.Email.Provider {
	case "sendgrid":
		sender, err = sendgrid.New(ctx, isConfig.Email.Config, isConfig.Email.SendGrid)
	case "smtp":
		sender, err = smtp.New(ctx, isConfig.Email.Config, isConfig.Email.SMTP)
	}
	if err != nil {
		return err
	}
	var data emails.Data
	data.Network.Name = isConfig.Email.Network.Name
	data.Network.IdentityServerURL = isConfig.Email.Network.IdentityServerURL
	data.Network.ConsoleURL = isConfig.Email.Network.ConsoleURL
	messageData := f(data)
	if messageData == nil {
		return nil
	}

	templates, err := is.getEmailTemplates(ctx)
	if err != nil {
		return err
	}
	message, err := templates.Render(messageData)
	if err != nil {
		return err
	}
	if sender == nil {
		log.FromContext(ctx).WithFields(log.Fields(
			"to", message.RecipientAddress,
			"subject", message.Subject,
			"body", message.TextBody,
		)).Warn("Could not send email without email provider")
		return nil
	}
	return sender.Send(message)
}

// SendUserEmail sends an email to the given user.
func (is *IdentityServer) SendUserEmail(ctx context.Context, userIDs *ttnpb.UserIdentifiers, makeMessage func(emails.Data) email.MessageData) error {
	var usr *ttnpb.User
	err := is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usr, err = store.GetUserStore(db).GetUser(ctx, userIDs, &types.FieldMask{
			Paths: []string{"name", "primary_email_address"},
		})
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return err
	}
	err = is.SendEmail(ctx, func(data emails.Data) email.MessageData {
		data.SetUser(usr)
		return makeMessage(data)
	})
	if err != nil {
		return err
	}
	return nil
}

// SendContactsEmail sends an email to the contacts of the given entity.
func (is *IdentityServer) SendContactsEmail(ctx context.Context, ids *ttnpb.EntityIdentifiers, makeMessage func(emails.Data) email.MessageData) error {
	var contacts []*ttnpb.ContactInfo
	err := is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		contacts, err = store.GetContactInfoStore(db).GetContactInfo(ctx, ids)
		return err
	})
	if err != nil {
		return err
	}
	for _, contactInfo := range contacts {
		if contactInfo.ContactMethod != ttnpb.CONTACT_METHOD_EMAIL {
			continue
		}
		err := is.SendEmail(ctx, func(data emails.Data) email.MessageData {
			data.SetEntity(ids)
			data.SetContact(contactInfo)
			return makeMessage(data)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
