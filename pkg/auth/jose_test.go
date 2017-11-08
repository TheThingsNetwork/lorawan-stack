// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/apikey"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestJOSEHeader(t *testing.T) {
	a := assertions.New(t)

	// apikey
	key, err := apikey.GenerateApplicationAPIKey("")
	a.So(err, should.BeNil)
	a.So(key, should.NotBeEmpty)

	header, err := JOSEHeader(key)
	a.So(err, should.BeNil)
	a.So(header, should.Resemble, &Header{
		Type:      apikey.Type,
		Algorithm: "secret",
	})
}
