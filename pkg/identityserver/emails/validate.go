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

// Validate is the validation email.
type Validate struct {
	Data
	ID    string
	Token string
}

// TemplateName returns the name of the template to use for this email.
func (Validate) TemplateName() string { return "validate" }

const validateSubject = `Please validate your contact info for {{.Network.Name}}`

const validateText = `Please validate your contact info for {{.Network.Name}}.

Your info will be used as contact for {{.Entity.Type}} "{{.Entity.ID}}".

Validate via web interface: {{ .Network.IdentityServerURL }}/validate?reference={{ .ID }}&token={{ .Token }}

Reference: {{.ID}}
Validation Token: {{.Token}}
`

// DefaultTemplates returns the default templates for this email.
func (Validate) DefaultTemplates() (subject, html, text string) {
	return validateSubject, "", validateText
}
