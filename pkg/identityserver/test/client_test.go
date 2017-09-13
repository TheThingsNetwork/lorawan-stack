// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func client() *types.DefaultClient {
	return &types.DefaultClient{
		ID:     "test-client",
		Secret: "123456",
		URI:    "/oauth/callback",
		Grants: types.Grants{Password: true, RefreshToken: true},
		Scope:  types.Scopes{Application: true},
	}
}

func TestShouldBeClient(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeClient(client(), client()), should.Equal, success)

	modified := client()
	modified.Created = time.Now()

	a.So(ShouldBeClient(modified, client()), should.NotEqual, success)
}

func TestShouldBeClientIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeClientIgnoringAutoFields(client(), client()), should.Equal, success)

	modified := client()
	modified.Secret = "foo"
	modified.Grants = types.Grants{}

	a.So(ShouldBeClientIgnoringAutoFields(modified, client()), should.NotEqual, success)
}
