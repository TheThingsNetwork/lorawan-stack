// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestJOSEEncoding(t *testing.T) {
	a := assertions.New(t)

	// Access Token
	{
		key, err := GenerateAccessToken("local")
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := Decode(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      typeToken,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: "local",
		})
	}

	// Application API Key
	{
		key, err := GenerateApplicationAPIKey("foo.issuer")
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := Decode(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      typeKey,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: "foo.issuer",
			Type:   typeApplication,
		})
	}

	// Gateway API Key
	{
		key, err := GenerateGatewayAPIKey("")
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := Decode(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      typeKey,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: "",
			Type:   typeGateway,
		})
	}
}
