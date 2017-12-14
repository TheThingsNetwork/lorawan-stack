// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

// EmailValidationHostname is the template field used to place the base URL.
const EmailValidationHostname = "hostname"

// EmailValidationToken is the template field used to place the validation token.
const EmailValidationToken = "token"

// EmailValidation is used to send emails asking to validate an email address.
func EmailValidation() *Template {
	return &Template{
		Name:    "Email validation",
		Subject: "Your email needs to be validated",
		Message: `<h1>Email verification</h1>

<p>
	You recently registered an account at
	<a href='https://thethingsnetwork.org'>The Things Network</a> using
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
