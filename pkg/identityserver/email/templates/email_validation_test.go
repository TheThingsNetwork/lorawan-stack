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

func TestEmailValidation(t *testing.T) {
	a := assertions.New(t)

	tmpl := EmailValidation{
		PublicURL:        "thethings.network",
		OrganizationName: "The Things Network",
		Token:            "bar",
	}

	subject, message, err := render(tmpl.GetName(), tmpl)
	a.So(err, should.BeNil)
	a.So(subject, should.Equal, "Your email needs to be validated")
	a.So(message, should.Equal, `<h1>Email verification</h1>

<p>
	You recently registered an account at
	<a href='thethings.network'>The Things Network</a> using
	this email address.
</p>

<p>
	Please activate
	your account by clicking the button below.
</p>

<p>
	<a class='button' href='thethings.network/api/v3/validate/bar'>Activate account</a>
</p>

<p class='extra'>
	If you did not register an account, you can ignore this e-mail.
</p>`)
}
