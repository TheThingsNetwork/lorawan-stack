// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

// AccountDeleted is the email template used when an account has been deleted.
type AccountDeleted struct {
	PublicURL        string
	OrganizationName string
	UserID           string
	Message          string
}

// Name implements Template.
func (t *AccountDeleted) GetName() string {
	return "Account deleted"
}

// Render implements Template.
func (t *AccountDeleted) Render() (string, string, error) {
	subject := "Your account has been deleted"
	message := `<h1>Account deleted</h1>

<p>
	Your account with ID {{.UserID}} at
	<a href='{{.PublicURL}}'>{{.OrganizationName}}</a>
	has been deleted by an admin.
</p>

{{if .Message}}
	<p>
		The admin has left the following message:
		<br>
		{{.Reason}}
	</p>
{{end}}`

	return render(subject, message, t)
}
