// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

// EmailValidation is the email template used to validate an email address.
type AccountCreation struct {
	PublicURL        string
	OrganizationName string
	Name             string
	UserID           string
	Password         string
	ValidationToken  string
}

// Name implements Template.
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
