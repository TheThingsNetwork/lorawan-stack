// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSubject(t *testing.T) {
	a := assertions.New(t)

	{
		sub := ApplicationSubject("foo")
		a.So(sub.String(), should.Equal, "application:foo")
		a.So(sub.Application(), should.Equal, "foo")
		a.So(sub.Gateway(), should.BeEmpty)
		a.So(sub.User(), should.BeEmpty)
	}

	{
		sub := GatewaySubject("foo")
		a.So(sub.String(), should.Equal, "gateway:foo")
		a.So(sub.Application(), should.BeEmpty)
		a.So(sub.Gateway(), should.Equal, "foo")
		a.So(sub.User(), should.BeEmpty)
	}

	{
		sub := UserSubject("foo")
		a.So(sub.String(), should.Equal, "user:foo")
		a.So(sub.Application(), should.BeEmpty)
		a.So(sub.Gateway(), should.BeEmpty)
		a.So(sub.User(), should.Equal, "foo")
	}
}
