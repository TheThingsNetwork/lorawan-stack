// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestKeysRSA(t *testing.T) {
	a := assertions.New(t)

	k := NewKeys("http://foo")

	key, err := rsa.GenerateKey(rand.Reader, 2014)
	a.So(err, should.BeNil)

	kid := "kid"

	_, err = k.GetPrivateKey(kid)
	a.So(err, should.NotBeNil)

	_, err = k.GetPublicKey(kid)
	a.So(err, should.NotBeNil)

	// rotate in new key
	err = k.Rotate(kid, key)
	a.So(err, should.BeNil)

	got, err := k.GetPrivateKey(kid)
	a.So(err, should.BeNil)
	a.So(got, should.Equal, key)

	gkey, err := k.GetPublicKey(kid)
	a.So(err, should.BeNil)
	a.So(gkey, should.Resemble, gkey)

	err = k.Rotate(kid, key)
	a.So(err, should.NotBeNil)
}

func TestManagerECDSA(t *testing.T) {
	a := assertions.New(t)

	k := &Keys{}

	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	a.So(err, should.BeNil)

	kid := "kid"

	// no key set, so return not found
	_, err = k.GetPrivateKey(kid)
	a.So(err, should.NotBeNil)

	_, err = k.GetPublicKey(kid)
	a.So(err, should.NotBeNil)

	err = k.Rotate(kid, key)
	a.So(err, should.BeNil)

	got, err := k.GetPrivateKey(kid)
	a.So(err, should.BeNil)
	a.So(got, should.Equal, key)

	gkey, err := k.GetPublicKey(kid)
	a.So(err, should.BeNil)

	a.So(gkey, should.Resemble, gkey)

	err = k.Rotate(kid, key)
	a.So(err, should.NotBeNil)
}

func TestManagerFromPEM(t *testing.T) {
	a := assertions.New(t)
	kid := "kid"

	pems := []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAzS36Huq9Hyk9W2pe3nzw+yzHwvDOVVw6XFG+023A+oQx7zOB
KQeup7Qz0bn/nAB7+Vt3Iu8/aElzg7/pNafA9xvPs6PsgXECVk2g+CUABSbRnq1p
s7MBBhZLsYzZbZ5ebkwbqcP/lzeO0dO5UTILAxqDmc7hs5PFuTqimaQIPq2HRPC5
JhYzHQnGbnBcPILh2Gh9cvorMTR9lSQFkUky8kYBlU4XGTxCgATf+AbQb6SXILd0
MuXqw7GM6T8jtNL0t9xFPEtSCJpj/GYFJZ1JB5gMfHlBgcqJ9vYZtwAvcV8JByHs
OEPrVBn9Q1z6QpNLtSl6MrZqTHaPM2bUtJmjVwIDAQABAoIBAQCZEk0A1cWEQuMI
mUHvoKyz5sOdVsPIgQb1KvM/jykifI84Umdwsc+GQ/VI6QeeXeofrTIjePQIHIw2
ZW1Z3y4h7Li233upUiMZOc72cbwjG8PVKrCqJMiFvwp3iooHstfmV5dnvtam/Qbq
2Zbu0XPPu+8kR2iw7XTcbLc0AmE4SCzmz/KfV6VmphBsAED4vMYN748g5ugLO8vO
/rIGopafC1ESx/k+ANxSBWqVUrOvjwepvQNDzeZPhPNQVKjrwte8d858AuR0tgsk
e1zetPyiqdKuS/ahic+X7ZT2PCJoGONXikx2LVSTeXshyZbkHHDd1SKGEeEN3tlX
ovwIVweBAoGBAPTDw/wpaVnWMBxc9BD6Xzmj3lqeQRPFloOWFv1xT0pLBKcOeVzG
6TXf7c9rWAuL2WGfYTyGIMrn/oEmivEZqoh6YZGT25f5jJxUmBanpAXvLgvTrMKs
nhic6n982IipHHs3xNWJ+UCAYYCD5TFYdVVKhEgHIpXuzQAmQeSszyRBAoGBANaZ
C86myKBsbmf0NpKTsKwKGWarnu7fZmBFSx0tRFoQu7jF1aeKWn48K0sH60vcLqdb
OeglvBTW8Z24jot6e1No5zjG6wJn141dzsXTBYMp+RPejXtCWcuwQP5lhgaRKSr6
ztbKIyyMTWd9GUs2SEOCRmcoHUmgLL5lsp+ZhwGXAoGBAN422TiWtDG+dmFZtq+v
T0K6VkW5BWYY7eQ7IFYqSA0v/GJajq4/XDzwNywnzYB2D/5EP3g+YYk1hGbmgiAP
6DYNvYT4UtYv1oubdZSj0BMfKZPNMjxvkPzRgUgLJV81AUmQwSAJKoR3yY1usWbS
Y3vyshPefnTWn8Ex+oLMrSbBAoGBAKyCDU7LHh5v6/TfCXudA/nYiIDTV4j4xyh+
q5pByF+Kcg1f45eyDXrKzZacQBcUYeCg4hTvOJmcDFDYiqYvCLKNcspehY7CgTGg
BldagmTlOdgyIJPES8EE58pZPHtM98YYJmvdxJbMFnEpzEp80WyLbiMAyUJlY3KO
+B96YF/zAoGAfkSN4SgkxYWVwikRLTQ7YhQxD/anNaa+y4QrmaS7ccZuQPXUN0KF
PdrF5e0X8uB76vq60osPk0R+41iZMu6gWcgL/rx5LqtCR9RJBS/vmlrQsKkDtwng
pfeKo3HLUYMyS8l55ppjahjP4nG2cvuayO/VaHUIJW6VoVn5VDZ4ukM=
-----END RSA PRIVATE KEY-----
`)
	k := &Keys{}
	err := k.RotateFromPEM(kid, pems)
	a.So(err, should.BeNil)

	_, err = k.GetPublicKey(kid)
	a.So(err, should.BeNil)

	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 1800,
			IssuedAt:  time.Now().Unix() - 1800,
			Subject:   ApplicationSubject("foo-app"),
			Issuer:    "account.thethingsnetwork.org",
		},
		User: "john-doe",
		Rights: []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_INFO,
			ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		},
	}

	str, err := k.Sign(claims)
	a.So(err, should.BeNil)
	a.So(str, should.NotBeEmpty)
}

func TestCheck(t *testing.T) {
	a := assertions.New(t)

	// ecdsa
	{
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		a.So(err, should.BeNil)

		a.So(checkPrivateKey(key), should.BeNil)
		a.So(checkPublicKey(&key.PublicKey), should.BeNil)

		a.So(checkPublicKey(key), should.NotBeNil)
		a.So(checkPrivateKey(key.PublicKey), should.NotBeNil)
	}

	// rsa
	{
		key, err := rsa.GenerateKey(rand.Reader, 2014)
		a.So(err, should.BeNil)

		a.So(checkPrivateKey(key), should.BeNil)
		a.So(checkPublicKey(&key.PublicKey), should.BeNil)

		a.So(checkPublicKey(key), should.NotBeNil)
		a.So(checkPrivateKey(key.PublicKey), should.NotBeNil)
	}

	// other stuff
	{
		a.So(checkPublicKey("foo"), should.NotBeNil)
		a.So(checkPrivateKey("bar"), should.NotBeNil)
	}
}
