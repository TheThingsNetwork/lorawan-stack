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

// AccountCreation is the template used when an admin creates an account in
// behalf of an user.
type AccountCreation struct {
	PublicURL        string
	OrganizationName string
	Name             string
	UserID           string
	Password         string
	ValidationToken  string
}

// GetName implements Template.
func (t *AccountCreation) GetName() string {
	return "Account creation on behalf of the user"
}

// Render implements Template.
func (t *AccountCreation) Render() (string, string, error) {
	subject := "You had been created an account in {{.OrganizationName}}"
	message := `<h1>Welcome{{if .ActivationToken}} {{.Name}}{{end}}</h1>

<p>
	You just got created an account at
	<a href='{{.PublicURL}}'>{{.OrganizationName}}</a> using
	this email address.
</p>

<p>
	Please note your account credentials
	as <b>{{.UserID}}</b> / <b>{{.Password}}</b>.
</p>

{{if .ActivationToken}}
	<p>
		Also please activate
		your account by clicking the button below.
	</p>

	<p>
		<a class='button' href='{{.PublicURL}}/api/v3/validate/{{.ActivationToken}}'>Activate account</a>
	</p>
{{end}}`

	return render(subject, message, t)
}
