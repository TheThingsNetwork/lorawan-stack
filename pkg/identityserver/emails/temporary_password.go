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

// TemporaryPassword is the email that is sent when users request a temporary password.
type TemporaryPassword struct {
	Data
	TemporaryPassword string
	TTL               time.Duration
}

// FormatTTL formats the TTL.
func (t TemporaryPassword) FormatTTL() string {
	return formatTTL(t.TTL)
}

// TemplateName returns the name of the template to use for this email.
func (TemporaryPassword) TemplateName() string { return "temporary_password" }

const temporaryPasswordSubject = `Your temporary password`

const temporaryPasswordText = `Dear {{.User.Name}},

A temporary password was requested for your user "{{.User.ID}}" on {{.Network.Name}}.

You can now go to {{ .Network.IdentityServerURL }}/update-password?user={{ .User.ID }}&current={{ .TemporaryPassword }} to change your password.

If you prefer to use the command-line interface, you can run the following command:

ttn-lw-cli users update-password --user-id {{.User.ID}} --old {{ .TemporaryPassword }} (add --revoke-all-access if you want to logout everywhere)

For more information on how to use the command-line interface, please refer to the documentation: {{ documentation_url "/getting-started/cli/" }}.

{{- if .TTL }}

Your temporary password expires {{ .FormatTTL }}, so change your password before then.
{{- end }}
`

// DefaultTemplates returns the default templates for this email.
func (TemporaryPassword) DefaultTemplates() (subject, html, text string) {
	return temporaryPasswordSubject, "", temporaryPasswordText
}
