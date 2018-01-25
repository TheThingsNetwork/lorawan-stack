// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sendgrid

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ email.Provider = new(SendGrid)

type template struct {
	name string
}

func (t *template) Name() string {
	return "template"
}

func (t *template) Render() (string, string, error) {
	return "hello", fmt.Sprintf("<b>%s!</b>", t.name), nil
}

func TestSendGridBuildEmail(t *testing.T) {
	a := assertions.New(t)

	sendgrid := New(
		test.GetLogger(t),
		Config{
			APIKey:      "API_KEY",
			SandboxMode: true,
			Name:        "Foo",
			From:        "foo@foo.local",
		},
	)

	a.So(sendgrid.client, should.NotBeNil)
	a.So(sendgrid.fromEmail, should.Resemble, mail.NewEmail("Foo", "foo@foo.local"))

	message, err := sendgrid.buildEmail("john@doe.com", &template{"john"})
	a.So(err, should.BeNil)
	a.So(message.From, should.Resemble, mail.NewEmail("Foo", "foo@foo.local"))
	a.So(message.Subject, should.Equal, "hello")
	a.So(message.Personalizations[0].To, should.Contain, mail.NewEmail("", "john@doe.com"))
	a.So(message.Content, should.HaveLength, 2)
	a.So(message.Content, should.Contain, mail.NewContent("text/html", "<b>john!</b>"))
	a.So(message.Content, should.Contain, mail.NewContent("text/plain", "*john!*"))
	a.So(*(message.MailSettings.SandboxMode.Enable), should.BeTrue)
}
