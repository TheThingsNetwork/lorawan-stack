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
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/email"
	"google.golang.org/grpc"
)

type testEmail struct {
	templateName string

	User struct {
		Name  string
		Email string
	}
}

func (e testEmail) Recipient() (name, address string) {
	return e.User.Name, e.User.Email
}

func (e testEmail) TemplateName() string { return e.templateName }

func (testEmail) DefaultTemplates() (subject, html, text string) {
	return "Welcome {{.User.Name}}", "HTML {{.User.Name}} {{.User.Email}}", "Text {{.User.Name}} {{.User.Email}}"
}

func TestGetEmailTemplates(t *testing.T) {
	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		for _, tc := range []struct {
			name    string
			message *email.Message
		}{
			{
				name: "Default",
				message: &email.Message{
					TemplateName:     "default",
					RecipientName:    "foo",
					RecipientAddress: "bar",
					Subject:          "Welcome foo",
					HTMLBody:         "HTML foo bar",
					TextBody:         "Text foo bar",
				},
			},
			{
				name: "Overridden",
				message: &email.Message{
					TemplateName:     "overridden",
					RecipientName:    "foo",
					RecipientAddress: "bar",
					Subject:          "Overridden subject foo",
					HTMLBody:         "Overridden HTML foo bar",
					TextBody:         "Overridden text foo bar",
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				a := assertions.New(t)
				email := testEmail{
					templateName: tc.message.TemplateName,
				}
				email.User.Name = tc.message.RecipientName
				email.User.Email = tc.message.RecipientAddress

				message, err := is.emailTemplates.Render(email)

				a.So(err, should.BeNil)
				a.So(message, should.Resemble, tc.message)
			})
		}
	})
}
