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

import "fmt"

// UserRequested is the email that is sent to admins when a user requests to join the network.
type UserRequested struct {
	Data
}

// ConsoleURL returns the URL to the user in the Console.
func (u UserRequested) ConsoleURL() string {
	return fmt.Sprintf("%s/admin/user-management/%s", u.Network.ConsoleURL, u.Entity.ID)
}

// TemplateName returns the name of the template to use for this email.
func (UserRequested) TemplateName() string { return "user_requested" }

const userRequestedSubject = `User registration of {{ .Entity.ID }} needs review`

const userRequestedText = `Dear {{ .User.Name }},

A new user "{{ .Entity.ID }}" was just registered on {{ .Network.Name }} and needs admin approval.

You can go to {{ .ConsoleURL }} to view and approve (or reject) the user.

If you prefer to use the command-line interface, you can run the following command:

ttn-lw-cli users set {{ .Entity.ID }} --state APPROVED (or --state REJECTED) --state-description "..."

For more information on how to use the command-line interface, please refer to the documentation: {{ documentation_url "/getting-started/cli/" }}.
`

// DefaultTemplates returns the default templates for this email.
func (UserRequested) DefaultTemplates() (subject, html, text string) {
	return userRequestedSubject, "", userRequestedText
}
