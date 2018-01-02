// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSettingsIsEmailAllowed(t *testing.T) {
	a := assertions.New(t)
	s := &IdentityServerSettings{}

	// all emails are allowed
	s.AllowedEmails = []string{}
	a.So(s.IsEmailAllowed("foo@foo.com"), should.BeTrue)
	a.So(s.IsEmailAllowed("foo@foofofofo.com"), should.BeTrue)

	// all emails are allowed
	s.AllowedEmails = []string{"*"}
	a.So(s.IsEmailAllowed("foo@foo.com"), should.BeTrue)
	a.So(s.IsEmailAllowed("foo@foofofofo.com"), should.BeTrue)

	// only emails ended in @ttn.org
	s.AllowedEmails = []string{"*@ttn.org"}
	a.So(s.IsEmailAllowed("foo@foo.com"), should.BeFalse)
	a.So(s.IsEmailAllowed("foo@foofofofo.com"), should.BeFalse)
	a.So(s.IsEmailAllowed("foo@ttn.org"), should.BeTrue)
	a.So(s.IsEmailAllowed("foo@TTN.org"), should.BeTrue)
}
