// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func client() *ttnpb.Client {
	return &ttnpb.Client{
		ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-client"},
		Secret:            "123456",
		RedirectURI:       "/oauth/callback",
		Grants:            []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE},
		Creator:           ttnpb.UserIdentifiers{UserID: "bob"},
	}
}

func TestShouldBeClient(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeClient(client(), client()), should.Equal, success)

	modified := client()
	modified.CreatedAt = time.Now()

	a.So(ShouldBeClient(modified, client()), should.NotEqual, success)
}

func TestShouldBeClientIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeClientIgnoringAutoFields(client(), client()), should.Equal, success)

	modified := client()
	modified.Description = "lol"

	a.So(ShouldBeClientIgnoringAutoFields(modified, client()), should.NotEqual, success)
}
