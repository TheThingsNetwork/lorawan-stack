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

// ClientDeleted is the email template used when an account has been deleted.
type ClientDeleted struct {
	PublicURL        string
	OrganizationName string
	ClientID         string
	Message          string
}

// GetName implements Template.
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
