// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestClaims(t *testing.T) {
	a := assertions.New(t)

	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 1800,
			IssuedAt:  time.Now().Unix() - 1800,
			Issuer:    "account.thethingsnetwork.org",
		},
		Subject: ApplicationSubject("foo"),
		Scope: []Scope{
			ApplicationInfo,
			ApplicationTrafficRead,
		},
	}

	a.So(claims.Valid(), should.BeNil)
	a.So(claims.HasScope(ApplicationInfo), should.BeTrue)
	a.So(claims.HasScope(ApplicationSettingsBasic), should.BeFalse)
	a.So(claims.HasScope(ApplicationInfo, ApplicationTrafficRead), should.BeTrue)
	a.So(claims.HasScope(ApplicationInfo, ApplicationDelete), should.BeFalse)
}

func TestSign(t *testing.T) {
	a := assertions.New(t)

	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 1800,
			IssuedAt:  time.Now().Unix() - 1800,
			Issuer:    "account.thethingsnetwork.org",
		},
		Subject: ApplicationSubject("foo"),
		Scope: []Scope{
			ApplicationInfo,
			ApplicationTrafficRead,
		},
	}

	// ECDSA512
	{
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		a.So(err, should.BeNil)

		token, err := claims.Sign(key)
		a.So(err, should.BeNil)
		a.So(token, should.NotBeEmpty)

		provider := &ConstProvider{
			Tokens: map[string]map[string]crypto.PublicKey{
				claims.Issuer: map[string]crypto.PublicKey{
					"": &key.PublicKey,
				},
			},
		}

		parsed, err := FromToken(provider, token)
		a.So(err, should.BeNil)
		a.So(parsed, should.Resemble, claims)
	}

	// RSA512
	{
		key, err := rsa.GenerateKey(rand.Reader, 2014)
		a.So(err, should.BeNil)

		token, err := claims.Sign(key)
		a.So(err, should.BeNil)
		a.So(token, should.NotBeEmpty)

		provider := &ConstProvider{
			Tokens: map[string]map[string]crypto.PublicKey{
				claims.Issuer: map[string]crypto.PublicKey{
					"": &key.PublicKey,
				},
			},
		}

		parsed, err := FromToken(provider, token)
		a.So(err, should.BeNil)
		a.So(parsed, should.Resemble, claims)
	}

	// ECDSA512 with kid
	{
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		a.So(err, should.BeNil)

		kid := "kid-123"
		token, err := claims.Sign(WithKID(key, kid))
		a.So(err, should.BeNil)
		a.So(token, should.NotBeEmpty)

		provider := &ConstProvider{
			Tokens: map[string]map[string]crypto.PublicKey{
				claims.Issuer: map[string]crypto.PublicKey{
					kid: &key.PublicKey,
				},
			},
		}

		parsed, err := FromToken(provider, token)
		a.So(err, should.BeNil)
		a.So(parsed, should.Resemble, claims)
	}

	// ECDSA512 with wrong kid
	{
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		a.So(err, should.BeNil)

		token, err := claims.Sign(WithKID(key, "kid"))
		a.So(err, should.BeNil)
		a.So(token, should.NotBeEmpty)

		provider := &ConstProvider{
			Tokens: map[string]map[string]crypto.PublicKey{
				claims.Issuer: map[string]crypto.PublicKey{
					"otherkid": &key.PublicKey,
				},
			},
		}

		_, err = FromToken(provider, token)
		a.So(err, should.NotBeNil)
	}
}
