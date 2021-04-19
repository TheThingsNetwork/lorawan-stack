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

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// CollaboratorChanged is the email that is sent when a collaborator is changed
type CollaboratorChanged struct {
	Data
	Collaborator ttnpb.Collaborator
}

// ConsoleURL returns the URL to the collaborator in the Console.
func (c CollaboratorChanged) ConsoleURL() string {
	return fmt.Sprintf("%s/%ss/%s/collaborators/%s/%s", c.Network.ConsoleURL, c.Entity.Type, c.Entity.ID, c.Collaborator.EntityType(), c.Collaborator.IDString())
}

// TemplateName returns the name of the template to use for this email.
func (CollaboratorChanged) TemplateName() string { return "collaborator_changed" }

const collaboratorChangedSubject = `A collaborator has been changed`

const collaboratorChangedBody = `Dear {{.User.Name}},

The collaborator "{{.Collaborator.IDString}}" of {{.Entity.Type}} "{{.Entity.ID}}" on {{.Network.Name}} now has the following rights:
{{range $right := .Collaborator.Rights}}
{{$right}} {{end}}

You can go to {{.ConsoleURL}} to view and edit this collaborator in the Console.

If you prefer to use the command-line interface, you can run the following commands to view or edit this collaborator:

ttn-lw-cli {{.Entity.Type}}s collaborators get --{{.Entity.Type}}-id {{.Entity.ID}} --{{.Collaborator.EntityType}}-id {{.Collaborator.IDString}}
ttn-lw-cli {{.Entity.Type}}s collaborators set --{{.Entity.Type}}-id {{.Entity.ID}} --{{.Collaborator.EntityType}}-id {{.Collaborator.IDString}}

For more information on how to use the command-line interface, please refer to the documentation: {{ documentation_url "/getting-started/cli/" }}.
`

// DefaultTemplates returns the default templates for this email.
func (CollaboratorChanged) DefaultTemplates() (subject, html, text string) {
	return collaboratorChangedSubject, "", collaboratorChangedBody
}
