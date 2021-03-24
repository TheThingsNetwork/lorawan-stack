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

package emails

import "time"

// LoginToken is the email that is sent when users request a login token.
type LoginToken struct {
	Data
	LoginToken string
	TTL        time.Duration
}

// FormatTTL formats the TTL.
func (t LoginToken) FormatTTL() string {
	return formatTTL(t.TTL)
}

// TemplateName returns the name of the template to use for this email.
func (LoginToken) TemplateName() string { return "login_token" }

const loginTokenSubject = `Your login token`

const loginTokenText = `Dear {{.User.Name}},

A login token was requested for your user "{{.User.ID}}" on {{.Network.Name}}.

You can now go to {{ .Network.IdentityServerURL }}/token-login?token={{ .LoginToken }} to log in.

{{- if .TTL }}

Your login token expires {{ .FormatTTL }}.
{{- end }}
`

// DefaultTemplates returns the default templates for this email.
func (LoginToken) DefaultTemplates() (subject, html, text string) {
	return loginTokenSubject, "", loginTokenText
}
