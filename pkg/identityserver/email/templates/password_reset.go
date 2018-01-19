// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

// EmailValidation is the email template used to validate an email address.
type PasswordReset struct {
	PublicURL        string
	OrganizationName string
	Password         string
}

// GetName implements Template.
func (t *PasswordReset) GetName() string {
	return "Password Reset"
}

// Render implements Template.
func (t *PasswordReset) Render() (string, string, error) {
	subject := "Your password has been reset"
	message := `<h1>Password reset</h1>

<p>
	Your password has been reset by a
	<a href='{{.PublicURL}}'>{{.OrganizationName}}</a> admin.
</p>

<p>
	Your new account's password is <b>{{.Password}}</b>
</p>`

	return render(subject, message, t)
}
