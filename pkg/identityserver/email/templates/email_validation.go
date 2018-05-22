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

func init() {
	templateName := new(EmailValidation).GetName()
	subject := "Your email needs to be validated"
	body := `<h1>Email verification</h1>

<p>
	You recently registered an account at
	<a href='{{.PublicURL}}'>{{.OrganizationName}}</a> using
	this email address.
</p>

<p>
	Please activate
	your account by clicking the button below.
</p>

<p>
	<a class='button' href='{{.PublicURL}}/api/v3/validate/{{.Token}}'>Activate account</a>
</p>

<p class='extra'>
	If you did not register an account, you can ignore this e-mail.
</p>`

	templates.Register(templateName, subject, body)
}

// EmailValidation is the email template used to validate an email address.
type EmailValidation struct {
	PublicURL        string
	OrganizationName string
	Token            string
}

// GetName implements Template.
func (t *EmailValidation) GetName() string {
	return "Email Validation"
}

// Render implements Template.
func (t *EmailValidation) Render() (string, string, error) {
	return render(t.GetName(), t)
}
