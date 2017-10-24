// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var (
	tenants = []string{
		"foo.bar.baz",
		"id.thethings.network",
		"a.very.long.tenant.id.that.is.really.long",
	}
)

func TestGenerateKey(t *testing.T) {
	a := assertions.New(t)

	// test good apps
	for _, tenant := range tenants {
		key, err := GenerateAPIKey(tenant)
		a.So(err, should.BeNil)

		ten, err := KeyTenant(key)
		a.So(err, should.BeNil)

		a.So(ten, should.Equal, tenant)
	}
}

func TestParseKey(t *testing.T) {
	a := assertions.New(t)

	p, err := marshal(payload{
		Issuer: "id.thethings.network",
	})
	a.So(err, should.BeNil)

	// bad payload
	{
		_, err := KeyTenant(header64 + ".invalid.secret")
		a.So(err, should.NotBeNil)
	}

	// bad secret (empty)
	{
		_, err = KeyTenant(header64 + "." + p + ".")
		a.So(err, should.NotBeNil)
	}

	// bad secret (none)
	{
		_, err = KeyTenant(header64 + "." + p)
		a.So(err, should.NotBeNil)
	}

	// bad issuer (none)
	{
		p, err := marshal(payload{})
		a.So(err, should.BeNil)

		_, err = KeyTenant(header64 + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (enc)
	{
		_, err = KeyTenant("invalid" + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (empty)
	{
		_, err = KeyTenant("." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (no values)
	{
		head, err := marshal(header{})
		a.So(err, should.BeNil)

		_, err = KeyTenant(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (no typ)
	{
		head, err := marshal(header{
			Alg: alg,
		})
		a.So(err, should.BeNil)

		_, err = KeyTenant(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (bad typ)
	{
		head, err := marshal(header{
			Type: "bad",
			Alg:  alg,
		})
		a.So(err, should.BeNil)

		_, err = KeyTenant(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (no alg)
	{
		head, err := marshal(header{
			Type: typ,
		})
		a.So(err, should.BeNil)

		_, err = KeyTenant(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (bad alg)
	{
		head, err := marshal(header{
			Type: typ,
			Alg:  "bad",
		})
		a.So(err, should.BeNil)

		_, err = KeyTenant(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}
}
