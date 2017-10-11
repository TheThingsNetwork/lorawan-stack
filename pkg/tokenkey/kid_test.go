// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestKeys(t *testing.T) {
	a := assertions.New(t)

	key, err := rsa.GenerateKey(rand.Reader, 2014)
	a.So(err, should.BeNil)

	kid := GetKID(key)
	a.So(kid, should.BeEmpty)

	k := "foo"
	withKID := WithKID(key, k)

	a.So(GetKID(withKID), should.Equal, k)

	var p *crypto.PrivateKey
	a.So(withKID, should.Implement, p)
}
