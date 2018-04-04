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

func (t *template) GetName() string {
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
