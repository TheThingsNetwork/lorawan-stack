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

package email_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/fetch"
)

var fetcher fetch.Interface

// globalData would be global data that may be used in every email.
type globalData struct {
	Network struct {
		Name              string
		IdentityServerURL string
		ConsoleURL        string
		// etc...
	}
	User struct {
		ID    string
		Name  string
		Email string
	}
}

func (g globalData) Recipient() (name, address string) {
	return g.User.Name, g.User.Email
}

// welcomeEmail is the data specifically for the welcome email.
type welcomeEmail struct {
	globalData
	ActivationToken string
}

// TemplateName returns the name of the template that the registry should search for.
func (welcome welcomeEmail) TemplateName() string { return "welcome" }

const welcomeSubject = `Welcome to {{.Network.Name}}`

const welcomeHTML = `<div style="styles to hide the pre-header-text">Welcome to {{.Network.Name}}, {{.User.Name}}!</div>
{{ template "header.html" . }}
<h1>Welcome to {{.Network.Name}}, {{.User.Name}}!</h1><br>
Please activate your account by visiting <a href="{{.Network.IdentityServerURL}}/activate/{{.ActivationToken}}">this link</a>.
{{ template "footer.html" . }}`

const welcomeText = `{{ template "header.txt" . }}

Welcome to {{.Network.Name}}, {{.User.Name}}!

Please activate your account by visiting {{.Network.IdentityServerURL}}/activate/{{.ActivationToken}}.

{{ template "footer.txt" . }}`

func (welcome welcomeEmail) DefaultTemplates() (subject, html, text string) {
	return welcomeSubject, welcomeHTML, welcomeText
}

func TestEmail(t *testing.T) {
	a := assertions.New(t)

	registry := email.NewTemplateRegistry(fetch.FromFilesystem("testdata"), "header.html", "footer.html", "header.txt", "footer.txt")

	data := welcomeEmail{}
	data.User.Name = "John Doe"
	data.User.Email = "john.doe@example.com"
	data.Network.Name = "The Things Network"
	data.Network.IdentityServerURL = "https://id.thethings.network"

	message, err := registry.Render(data)
	a.So(err, should.BeNil)
	if a.So(message, should.NotBeNil) {
		a.So(message.Subject, should.Equal, "Welcome to The Things Network")
		a.So(message.HTMLBody, should.ContainSubstring, `<div class="header">`)
		a.So(message.HTMLBody, should.ContainSubstring, "Welcome to The Things Network, John Doe!")
		a.So(message.HTMLBody, should.ContainSubstring, `<div class="footer">`)
		a.So(message.TextBody, should.ContainSubstring, "==================")
		a.So(message.TextBody, should.ContainSubstring, "Welcome to The Things Network, John Doe!")
	}
}

func Example() {
	// The email sender can be Sendgrid, SMTP, ...
	var sender email.Sender

	// This can fetch templates from the filesystem, github, S3, ...
	registry := email.NewTemplateRegistry(fetcher)

	data := welcomeEmail{}
	data.User.Name = "John Doe"
	data.User.Email = "john.doe@example.com"
	data.Network.Name = "The Things Network"
	data.Network.IdentityServerURL = "https://id.thethings.network"

	// The first time you render an email, the template will also be compiled.
	// Any later changes to the template will not be picked up.
	// The compiled template will render into an email that is ready to be sent.
	message, err := registry.Render(data)
	if err != nil {
		return // error rendering the message
	}

	err = sender.Send(message)
	if err != nil {
		return // error sending the message
	}

	// done!
}
