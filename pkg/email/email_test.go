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
	"fmt"
	"os"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// welcomeEmailData is the data specifically for the welcome email.
type welcomeEmailData struct {
	email.TemplateData
	ActivationToken string
}

func TestEmail(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)

	registry := email.NewTemplateRegistry()

	welcomeEmailTemplate, err := email.NewTemplateFS(
		os.DirFS("testdata"), ttnpb.NotificationType_UNKNOWN,
		email.FSTemplate{
			SubjectTemplate:      "Welcome to {{ .Network.Name }}",
			HTMLTemplateBaseFile: "base.html",
			HTMLTemplateFile:     "welcome.html",
			TextTemplateBaseFile: "base.txt",
			TextTemplateFile:     "welcome.txt",
			IncludePatterns:      []string{"header.html", "footer.html", "header.txt", "footer.txt"},
		},
	)
	a.So(err, should.BeNil)

	registry.RegisterTemplate(welcomeEmailTemplate)
	a.So(registry.RegisteredTemplates(), should.Contain, "UNKNOWN")
	returnedTemplate := registry.GetTemplate(ctx, ttnpb.NotificationType_UNKNOWN)

	for i, template := range []*email.Template{welcomeEmailTemplate, returnedTemplate} {
		template := template
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)

			message, err := template.Execute(&welcomeEmailData{
				TemplateData: email.NewTemplateData(&email.NetworkConfig{
					Name:              "The Things Network",
					IdentityServerURL: "https://eu1.cloud.thethings.network/oauth",
					ConsoleURL:        "https://console.cloud.thethings.network",
				}, &ttnpb.User{
					Name:                "John Doe",
					PrimaryEmailAddress: "john.doe@example.com",
				}),
			})

			if a.So(err, should.BeNil) && a.So(message, should.NotBeNil) {
				a.So(message.Subject, should.Equal, "Welcome to The Things Network")
				a.So(message.HTMLBody, should.ContainSubstring, `<div class="header">`)
				a.So(message.HTMLBody, should.ContainSubstring, "Welcome to The Things Network, John Doe!")
				a.So(message.HTMLBody, should.ContainSubstring, `<div class="footer">`)
				a.So(message.TextBody, should.ContainSubstring, "==================")
				a.So(message.TextBody, should.ContainSubstring, "Welcome to The Things Network, John Doe!")
			}
		})
	}
}
