// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

// EmailValidation is the email template used to validate an email address.
type EmailValidation struct {
	PublicURL        string
	OrganizationName string
	Token            string
}

// Name implements Template.
func (t *EmailValidation) Name() string {
	return "Email Validation"
}

// Render implements Template.
func (t *EmailValidation) Render() (string, string, error) {
	subject := "Your email needs to be validated"
	message := `<h1>Email verification</h1>

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

	return render(subject, message, t)
}
