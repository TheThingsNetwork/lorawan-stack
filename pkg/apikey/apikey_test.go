// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var (
	ids = []string{
		"foo.bar.baz",
		"id.thethings.network",
		"a.very.long.tenant.id.that.is.really.long",
	}
)

func TestGenerateKey(t *testing.T) {
	a := assertions.New(t)

	for _, id := range ids {
		key, err := GenerateApplicationAPIKey(id)
		a.So(err, should.BeNil)
		payload, err := KeyPayload(key)
		a.So(err, should.BeNil)
		a.So(payload, should.Resemble, &Payload{
			Issuer: id,
			Type:   TypeApplication,
		})

		key, err = GenerateGatewayAPIKey(id)
		a.So(err, should.BeNil)
		payload, err = KeyPayload(key)
		a.So(err, should.BeNil)
		a.So(payload, should.Resemble, &Payload{
			Issuer: id,
			Type:   TypeGateway,
		})
	}
}

func TestParseKey(t *testing.T) {
	a := assertions.New(t)

	p, err := marshal(Payload{
		Issuer: "id.thethings.network",
		Type:   TypeApplication,
	})
	a.So(err, should.BeNil)

	// bad payload
	{
		_, err := KeyPayload(header64 + ".invalid.secret")
		a.So(err, should.NotBeNil)
	}

	// bad secret (empty)
	{
		_, err = KeyPayload(header64 + "." + p + ".")
		a.So(err, should.NotBeNil)
	}

	// bad secret (none)
	{
		_, err = KeyPayload(header64 + "." + p)
		a.So(err, should.NotBeNil)
	}

	// bad issuer (none)
	{
		p, err := marshal(Payload{
			Type: TypeApplication,
		})
		a.So(err, should.BeNil)

		_, err = KeyPayload(header64 + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad type (none)
	{
		p, err := marshal(Payload{
			Issuer: "foo",
		})
		a.So(err, should.BeNil)

		_, err = KeyPayload(header64 + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (enc)
	{
		_, err = KeyPayload("invalid" + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (empty)
	{
		_, err = KeyPayload("." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (no values)
	{
		head, err := marshal(Header{})
		a.So(err, should.BeNil)

		_, err = KeyPayload(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (no typ)
	{
		head, err := marshal(Header{
			Alg: alg,
		})
		a.So(err, should.BeNil)

		_, err = KeyPayload(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (bad typ)
	{
		head, err := marshal(Header{
			Type: "bad",
			Alg:  alg,
		})
		a.So(err, should.BeNil)

		_, err = KeyPayload(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (no alg)
	{
		head, err := marshal(Header{
			Type: Type,
		})
		a.So(err, should.BeNil)

		_, err = KeyPayload(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}

	// bad header (bad alg)
	{
		head, err := marshal(Header{
			Type: Type,
			Alg:  "bad",
		})
		a.So(err, should.BeNil)

		_, err = KeyPayload(head + "." + p + ".secret")
		a.So(err, should.NotBeNil)
	}
}
