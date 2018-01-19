// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

// ClientDeleted is the email template used when an account has been deleted.
type ClientDeleted struct {
	PublicURL        string
	OrganizationName string
	ClientID         string
	Message          string
}

// Name implements Template.
func (t *ClientDeleted) GetName() string {
	return "Client deleted"
}

// Render implements Template.
func (t *ClientDeleted) Render() (string, string, error) {
	subject := "Your third-party client was deleted"
	message := `<h1>Client deleted</h1>

<p>
	Your third-party client with ID {{.ClientID}} at
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
