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
	templateName := new(PasswordReset).GetName()
	subject := "Your password has been reset"
	body := `<h1>Password reset</h1>

<p>
	Your password has been reset by a
	<a href='{{.PublicURL}}'>{{.OrganizationName}}</a> admin.
</p>

<p>
	Your new account's password is <b>{{.Password}}</b>
</p>`

	templates.Register(templateName, subject, body)
}

// PasswordReset is the email template used to inform an user that an admin has
// reset his account password.
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
	return render(t.GetName(), t)
}
