// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestClientScope(t *testing.T) {
	a := assertions.New(t)

	{
		scope := ClientScope{}

		value, err := scope.Value()
		a.So(err, should.BeNil)
		if str, ok := value.(string); a.So(ok, should.BeTrue) {
			a.So(str, should.BeEmpty)
		}

		scope.Application = true
		scope.Profile = true
		value, err = scope.Value()
		a.So(err, should.BeNil)
		if str, ok := value.(string); a.So(ok, should.BeTrue) {
			a.So(str, should.Equal, "0,1")
		}
	}

	{
		scope := &ClientScope{}

		src := "0"
		err := scope.Scan(src)
		a.So(err, should.BeNil)
		a.So(scope, should.Resemble, &ClientScope{Application: true})

		src = "0,1,2,3,4"
		err = scope.Scan(src)
		a.So(err, should.BeNil)
		a.So(scope, should.Resemble, &ClientScope{Application: true, Profile: true})

		data := 10
		err = scope.Scan(data)
		a.So(err, should.NotBeNil)
	}
}

func TestClientGrants(t *testing.T) {
	a := assertions.New(t)

	{
		grants := ClientGrants{}

		value, err := grants.Value()
		a.So(err, should.BeNil)
		if str, ok := value.(string); a.So(ok, should.BeTrue) {
			a.So(str, should.BeEmpty)
		}

		grants.AuthorizationCode = true
		grants.Password = true
		grants.RefreshToken = true
		value, err = grants.Value()
		a.So(err, should.BeNil)
		if str, ok := value.(string); a.So(ok, should.BeTrue) {
			a.So(str, should.Equal, "0,1,2")
		}
	}

	{
		grants := &ClientGrants{}

		src := "0"
		err := grants.Scan(src)
		a.So(err, should.BeNil)
		a.So(grants, should.Resemble, &ClientGrants{AuthorizationCode: true})

		src = "0,1,2,3,4"
		err = grants.Scan(src)
		a.So(err, should.BeNil)
		a.So(grants, should.Resemble, &ClientGrants{
			AuthorizationCode: true,
			Password:          true,
			RefreshToken:      true,
		})

		data := 10
		err = grants.Scan(data)
		a.So(err, should.NotBeNil)
	}
}
