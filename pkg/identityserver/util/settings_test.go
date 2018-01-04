// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package util

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestIsEmailAllowed(t *testing.T) {
	a := assertions.New(t)

	var allowedEmails []string

	// all emails are allowed
	allowedEmails = []string{}
	a.So(IsEmailAllowed("foo@foo.com", allowedEmails), should.BeTrue)
	a.So(IsEmailAllowed("foo@foofofofo.com", allowedEmails), should.BeTrue)

	// all emails are allowed
	allowedEmails = []string{"*"}
	a.So(IsEmailAllowed("foo@foo.com", allowedEmails), should.BeTrue)
	a.So(IsEmailAllowed("foo@foofofofo.com", allowedEmails), should.BeTrue)

	// only emails ended in @ttn.org
	allowedEmails = []string{"*@ttn.org"}
	a.So(IsEmailAllowed("foo@foo.com", allowedEmails), should.BeFalse)
	a.So(IsEmailAllowed("foo@foofofofo.com", allowedEmails), should.BeFalse)
	a.So(IsEmailAllowed("foo@ttn.org", allowedEmails), should.BeTrue)
	a.So(IsEmailAllowed("foo@TTN.org", allowedEmails), should.BeTrue)
}

func TestIsIDAllowed(t *testing.T) {
	a := assertions.New(t)

	var blacklistedIDs []string

	// all ids are allowed
	blacklistedIDs = nil
	a.So(IsIDAllowed("foobar", blacklistedIDs), should.BeTrue)
	a.So(IsIDAllowed("admin", blacklistedIDs), should.BeTrue)
	blacklistedIDs = []string{}
	a.So(IsIDAllowed("foobar", blacklistedIDs), should.BeTrue)
	a.So(IsIDAllowed("admin", blacklistedIDs), should.BeTrue)

	// `admin` is blacklisted
	blacklistedIDs = []string{"admin"}
	a.So(IsIDAllowed("foobar", blacklistedIDs), should.BeTrue)
	a.So(IsIDAllowed("admin", blacklistedIDs), should.BeFalse)
}
