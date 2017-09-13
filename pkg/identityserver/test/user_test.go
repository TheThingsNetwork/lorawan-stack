// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func user() *types.DefaultUser {
	return &types.DefaultUser{
		Username: "alice",
		Password: "123456",
		Email:    "alice@alice.com",
	}
}

func TestShouldBeUser(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeUser(user(), user()), should.Equal, success)

	modified := user()
	modified.Joined = time.Now()

	a.So(ShouldBeUser(modified, user()), should.NotEqual, success)
}

func TestShouldBeUserIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeUserIgnoringAutoFields(user(), user()), should.Equal, success)

	modified := user()
	modified.Password = "foo"

	a.So(ShouldBeUserIgnoringAutoFields(modified, user()), should.NotEqual, success)
}
