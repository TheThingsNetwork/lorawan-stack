// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// ClientRequested is the email that is sent to admins when a user requests an OAuth client.
type ClientRequested struct {
	Data
	Client       *ttnpb.Client
	Collaborator *ttnpb.OrganizationOrUserIdentifiers
}

// TemplateName returns the name of the template to use for this email.
func (ClientRequested) TemplateName() string { return "client_requested" }

const clientRequestedSubject = `OAuth client registration of {{ .Entity.ID }} needs review`

const clientRequestedText = `Dear {{ .User.Name }},

A new OAuth client "{{ .Entity.ID }}" was just registered by {{ .Collaborator.EntityType }} "{{ .Collaborator.IDString }}" on {{ .Network.Name }} and needs admin approval.

Name: {{ .Client.Name }}
Description: {{ .Client.Description }}

Grants:
{{- with .Client.Grants }}
{{- range . }}
- {{ . }}
{{- end }}
{{- else }} (none)
{{- end }}

Rights:
{{- with .Client.Rights }}
{{- range . }}
- {{ . }}
{{- end }}
{{- else }} (none)
{{- end }}

Redirect URIs:
{{- with .Client.RedirectURIs }}
{{- range . }}
- {{ . }}
{{- end }}
{{- else }} (none)
{{- end }}

Logout Redirect URIs:
{{- with .Client.LogoutRedirectURIs }}
{{- range . }}
- {{ . }}
{{- end }}
{{- else }} (none)
{{- end }}

You can use the command-line interface to approve (or reject) the OAuth client:

ttn-lw-cli clients set {{ .Entity.ID }} --state APPROVED (or --state REJECTED) --state-description "..."

For more information on how to use the command-line interface, please refer to the documentation: {{ documentation_url "/getting-started/cli/" }}.
`

// DefaultTemplates returns the default templates for this email.
func (ClientRequested) DefaultTemplates() (subject, html, text string) {
	return clientRequestedSubject, "", clientRequestedText
}
