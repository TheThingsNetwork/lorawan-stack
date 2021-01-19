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

// Invitation is the email that is sent when a user is invited to the network.
type Invitation struct {
	Data
	InvitationToken string
	TTL             time.Duration
}

// FormatTTL formats the TTL.
func (i Invitation) FormatTTL() string {
	return formatTTL(i.TTL)
}

// TemplateName returns the name of the template to use for this email.
func (Invitation) TemplateName() string { return "invitation" }

const invitationSubject = `Invitation to join {{ .Network.Name }}`

const invitationText = `Hello,

You have been invited to join {{ .Network.Name }}.

You can now go to {{ .Network.IdentityServerURL }}/register?invitation_token={{ .InvitationToken }} to register your user.

If you prefer to use the command-line interface, you can add "--invitation-token {{ .InvitationToken }}" when running the "ttn-lw-cli users create" command.

{{- if .TTL }}

Your invitation expires {{ .FormatTTL }}, so register your user before then.
{{- end }}

After successful registration, you can go to {{ .Network.ConsoleURL }} to start adding devices and gateways.

For more information on how how to get started, please refer to the documentation: {{ documentation_url "/getting-started/" }}.
`

// DefaultTemplates returns the default templates for this email.
func (Invitation) DefaultTemplates() (subject, html, text string) {
	return invitationSubject, "", invitationText
}
