// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ Provider = new(ConstProvider)

func TestConstProvider(t *testing.T) {
	a := assertions.New(t)

	iss := "identity.thethings.network"
	kid := ""

	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	a.So(err, should.BeNil)

	provider := &ConstProvider{
		Tokens: map[string]map[string]crypto.PublicKey{
			iss: map[string]crypto.PublicKey{
				kid: &key.PublicKey,
			},
		},
	}

	k, err := provider.TokenKey(iss, kid)
	a.So(err, should.BeNil)
	a.So(k, should.Resemble, &key.PublicKey)

	_, err = provider.TokenKey("wrong", kid)
	a.So(err, should.Equal, ErrUnknownIdentityServer)

	_, err = provider.TokenKey(iss, "wrong")
	a.So(err, should.Equal, ErrUnknownKID)
}
