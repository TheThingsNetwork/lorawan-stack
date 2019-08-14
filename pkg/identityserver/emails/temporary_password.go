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

// TemporaryPassword is the email that is sent when users request a temporary password.
type TemporaryPassword struct {
	Data
	TemporaryPassword string
}

// TemplateName returns the name of the template to use for this email.
func (TemporaryPassword) TemplateName() string { return "temporary_password" }

const temporaryPasswordSubject = `Your temporary password`

const temporaryPasswordText = `Dear {{.User.Name}},

A temporary password was requested for your user "{{.User.ID}}" on {{.Network.Name}}.

This temporary password can only be used once, and only to change the password of your account.

Temporary Password: {{.TemporaryPassword}}

If you wish to change the password using web interface, follow the link below:

{{ .Network.IdentityServerURL }}/update-password?user={{ .User.ID }}&current={{ .TemporaryPassword }}
`

// DefaultTemplates returns the default templates for this email.
func (TemporaryPassword) DefaultTemplates() (subject, html, text string) {
	return temporaryPasswordSubject, "", temporaryPasswordText
}
