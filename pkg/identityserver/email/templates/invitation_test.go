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
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestInvitation(t *testing.T) {
	a := assertions.New(t)

	tmpl := Invitation{
		PublicURL:        "thethings.network",
		OrganizationName: "The Things Network",
		WebUIURL:         "foo",
		Token:            "bar",
	}

	subject, message, err := render(tmpl.GetName(), tmpl)
	a.So(err, should.BeNil)
	a.So(subject, should.Equal, "You had been invited to join The Things Network!")
	a.So(message, should.Equal, `<h1>Invitation</h1>

<p>
	You just got invited to create an account
	at <a href='thethings.network'>The Things Network</a>.
</p>

<p>
	You can create your account by
	clicking the button below.
</p>

<p>
	<a class='button' href='foo/register?token=bar'>Create account</a>
</p>`)
}
