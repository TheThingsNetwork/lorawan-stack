// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var (
	tenant = "id.tenant.example.net"
	apps   = []string{
		"aa",
		"ab",
		"app-foo",
		"app-with-a-very-long-id-but-still-ok",
	}
)

func TestGenerateKey(t *testing.T) {
	a := assertions.New(t)

	// test good apps
	for _, app := range apps {
		key := GenerateApplicationAPIKey(tenant, app)
		info, err := DecodeKey(key)
		a.So(err, should.BeNil)

		a.So(info.ID, should.Equal, app)
		a.So(info.Tenant, should.Equal, tenant)
		a.So(info.Type, should.Equal, Application)
	}

}

func TestGenerateKeyInvalidKey(t *testing.T) {
	a := assertions.New(t)

	key := GenerateApplicationAPIKey(tenant, "app-with-a-very-long-id-that-is-definately-too-long-for-us")
	a.So(key, should.BeEmpty)
}

func TestGenerateKeyInvalidKeyType(t *testing.T) {
	a := assertions.New(t)

	key := generateKey(Invalid, tenant, "app-with-a-very-long-id-that-is-definately-too-long-for-us")
	a.So(key, should.BeEmpty)
}
