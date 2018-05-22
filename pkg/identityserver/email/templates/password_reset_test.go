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

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPasswordReset(t *testing.T) {
	a := assertions.New(t)

	tmpl := PasswordReset{
		PublicURL:        "thethings.network",
		OrganizationName: "The Things Network",
		Password:         "qwerty",
	}

	subject, message, err := render(tmpl.GetName(), tmpl)
	a.So(err, should.BeNil)
	a.So(subject, should.Equal, "Your password has been reset")
	a.So(message, should.Equal, `<h1>Password reset</h1>

<p>
	Your password has been reset by a
	<a href='thethings.network'>The Things Network</a> admin.
</p>

<p>
	Your new account's password is <b>qwerty</b>
</p>`)
}
