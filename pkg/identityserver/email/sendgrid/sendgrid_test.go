// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sendgrid

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/templates"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ email.Provider = new(SendGrid)

func TestSendGrid(t *testing.T) {
	a := assertions.New(t)

	tmpl := &templates.Template{
		Subject: "Hi",
		Message: "<b>{{.name}}!</b>",
	}

	sendgrid := New(
		test.GetLogger(t),
		"API_KEY",
		SandboxMode(true),
		SenderAddress("Foo", "foo@foo.local"),
	)

	a.So(sendgrid.client, should.NotBeNil)
	a.So(sendgrid.sandboxMode, should.BeTrue)
	a.So(sendgrid.fromEmail, should.Resemble, mail.NewEmail("Foo", "foo@foo.local"))

	message, err := sendgrid.buildEmail("john@doe.com", tmpl, map[string]interface{}{
		"name": "john",
	})
	a.So(err, should.BeNil)
	a.So(message.From, should.Resemble, mail.NewEmail("Foo", "foo@foo.local"))
	a.So(message.Subject, should.Equal, tmpl.Subject)
	a.So(message.Personalizations[0].To, should.Contain, mail.NewEmail("", "john@doe.com"))
	a.So(message.Content, should.HaveLength, 2)
	a.So(message.Content, should.Contain, mail.NewContent("text/html", "<b>john!</b>"))
	a.So(message.Content, should.Contain, mail.NewContent("text/plain", "*john!*"))
	a.So(*(message.MailSettings.SandboxMode.Enable), should.BeTrue)
}
