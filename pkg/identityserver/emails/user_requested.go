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

// UserRequested is the email that is sent to admins when a user requests to join the network.
type UserRequested struct {
	Data
}

// TemplateName returns the name of the template to use for this email.
func (UserRequested) TemplateName() string { return "user_requested" }

const userRequestedSubject = `User {{ .Entity.ID }} is waiting for approval`

const userRequestedText = `Dear {{ .User.Name }},

User "{{ .Entity.ID }}" wants to join {{ .Network.Name }}.

You can approve or reject them in the Console or using the Command-line interface.

You can read how exactly to do this in the user management guide:
https://thethingsstack.io/latest/getting-started/user-management/
`

// DefaultTemplates returns the default templates for this email.
func (UserRequested) DefaultTemplates() (subject, html, text string) {
	return userRequestedSubject, "", userRequestedText
}
