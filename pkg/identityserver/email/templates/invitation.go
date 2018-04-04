// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package templates

// Invitation is the email template used to notify a person that has been invited
// to register an account.
type Invitation struct {
	PublicURL        string
	OrganizationName string
	WebUIURL         string
	Token            string
}

// GetName implements Template.
func (t *Invitation) GetName() string {
	return "Invitation"
}

// Render implements Template.
func (t *Invitation) Render() (string, string, error) {
	subject := "You had been invited to join {{.OrganizationName}}!"
	message := `<h1>Invitation</h1>

<p>
	You just got invited to create an account
	at <a href='{{.PublicURL}}'>{{.OrganizationName}}</a>.
</p>

<p>
	You can create your account by
	clicking the button below.
</p>

<p>
	<a class='button' href='{{.WebUIURL}}/register?token={{.Token}}'>Create account</a>
</p>`

	return render(subject, message, t)
}
