// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

const (
	// EmailValidationDisplayName is the template field used to place the tenant display name.
	EmailValidationDisplayName = "display_name"

	// EmailValidationHomeURL is the template field used to place the tenant home url.
	EmailValidationHomeURL = "home_url"

	// EmailValidationHostname is the template field used to place the base URL.
	EmailValidationHostname = "hostname"

	// EmailValidationToken is the template field used to place the validation token.
	EmailValidationToken = "token"
)

// EmailValidation is used to send emails asking to validate an email address.
func EmailValidation() *Template {
	return &Template{
		Name:    "Email validation",
		Subject: "Your email needs to be validated",
		Message: `<h1>Email verification</h1>

<p>
	You recently registered an account at
	<a href='{{.home_url}}'>{{.display_name}}</a> using
	this email address.
</p>

<p>
	Please activate
	your account by clicking the button below.
</p>

<p>
	<a class='button' href='{{.hostname}}/validate/{{.token}}'>Activate account</a>
</p>

<p class='extra'>
	If you did not register an account, you can ignore this e-mail.
</p>`,
	}
}
