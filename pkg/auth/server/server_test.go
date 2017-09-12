// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestGetTokenKeyRSA(t *testing.T) {
	a := assertions.New(t)

	keys := new(auth.Keys)
	server := New(keys)

	key, err := rsa.GenerateKey(rand.Reader, 2014)
	a.So(err, should.BeNil)
	kid := "kid"

	err = keys.Rotate(kid, key)
	a.So(err, should.BeNil)

	resp, err := server.GetTokenKey(context.Background(), &ttnpb.TokenKeyRequest{
		KID: kid,
	})
	a.So(err, should.BeNil)
	a.So(resp.GetKID(), should.Equal, kid)
	a.So(resp.GetPublicKey(), should.ContainSubstring, "RSA PUBLIC KEY")
	a.So(resp.GetAlgorithm(), should.Equal, "RS512")

	_, err = server.GetTokenKey(context.Background(), &ttnpb.TokenKeyRequest{})
	a.So(err, should.NotBeNil)
}

func TestGetTokenKeyECDSA(t *testing.T) {
	a := assertions.New(t)

	keys := new(auth.Keys)
	server := New(keys)

	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	a.So(err, should.BeNil)
	kid := "kid"

	err = keys.Rotate(kid, key)
	a.So(err, should.BeNil)

	resp, err := server.GetTokenKey(context.Background(), &ttnpb.TokenKeyRequest{
		KID: kid,
	})
	a.So(err, should.BeNil)
	a.So(resp.GetKID(), should.Equal, kid)
	a.So(resp.GetPublicKey(), should.ContainSubstring, "EC PUBLIC KEY")
	a.So(resp.GetAlgorithm(), should.Equal, "ES512")

	_, err = server.GetTokenKey(context.Background(), &ttnpb.TokenKeyRequest{})
	a.So(err, should.NotBeNil)
}
