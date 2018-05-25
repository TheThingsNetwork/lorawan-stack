// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func BenchmarkKeyDecoding(b *testing.B) {
	key, _ := GenerateApplicationAPIKey("foo.issuer")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		DecodeTokenOrKey(key)
	}
}

func TestJOSEEncoding(t *testing.T) {
	a := assertions.New(t)

	// Access Token
	{
		key, err := GenerateAccessToken("local")
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := DecodeTokenOrKey(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      Token,
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

		header, payload, err := DecodeTokenOrKey(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      Key,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: "foo.issuer",
			Type:   ApplicationKey,
		})
	}

	// Gateway API Key
	{
		key, err := GenerateGatewayAPIKey("")
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := DecodeTokenOrKey(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      Key,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: "",
			Type:   GatewayKey,
		})
	}

	// User API Key
	{
		key, err := GenerateUserAPIKey("")
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := DecodeTokenOrKey(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      Key,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: "",
			Type:   UserKey,
		})
	}

	// Organization API Key
	{
		issuer := "foo.thethingsnetwork.org"

		key, err := GenerateOrganizationAPIKey(issuer)
		a.So(err, should.BeNil)
		a.So(key, should.NotBeEmpty)

		header, payload, err := DecodeTokenOrKey(key)
		a.So(err, should.BeNil)
		a.So(header, should.Resemble, &Header{
			Type:      Key,
			Algorithm: alg,
		})
		a.So(payload, should.Resemble, &Payload{
			Issuer: issuer,
			Type:   OrganizationKey,
		})
	}
}
