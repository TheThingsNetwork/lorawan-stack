// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// ErrUnsupportedSigningMethod occurs when trying to sign or verify  the claims using an unsupported signing method.
	ErrUnsupportedSigningMethod = errors.New("Signing method not supported")
)

// Claims is the type of claims used for the things network authentication.
// It consists of the standard JWT claims and a couple of additional fields.
type Claims struct {
	jwt.StandardClaims

	// Client is the client this token was created for.
	Client string `json:"cid,omitempty"`

	// Scope is the list of actions this token has access to.
	Scope ttnpb.Scope `json:"scope"`
}

// Username returns the username of the user this token is for, or the empty string if it is not for a user.
func (c *Claims) Username() string {
	return c.Scope.Username()
}

// ApplicationID returns the application ID  of the application this token is for, or the empty string if it is not for an application.
func (c *Claims) ApplicationID() string {
	return c.Scope.ApplicationID()
}

// GatewayID returns the gateway ID  of the gateway this token is for, or the empty string if it is not for a gateway.
func (c *Claims) GatewayID() string {
	return c.Scope.GatewayID()
}

// HasRights checks wether or not the provided right is included in the tokens scope. It will only return true if all the provided rights are
// included in the token..
func (c *Claims) HasRights(rights ...ttnpb.Right) bool {
	return c.Scope.HasRights(rights...)
}

// Sign signs the claims using the provided signing method and returns the corresponding JWT.
// The signing method is determined from the type of the private key provided.
// Currently ECDSA and RSA are supported.
func (c *Claims) Sign(privateKey crypto.PrivateKey) (string, error) {
	var method jwt.SigningMethod

	key := privateKey
	kid := ""

	// set the kid if it is there
	if w, ok := privateKey.(*PrivateKeyWithKID); ok {
		kid = w.KID
		key = w.PrivateKey
	}

	switch key.(type) {
	case *rsa.PrivateKey:
		method = jwt.SigningMethodRS512
	case *ecdsa.PrivateKey:
		method = jwt.SigningMethodES512
	default:
		return "", ErrUnsupportedSigningMethod
	}

	builder := jwt.NewWithClaims(method, c)
	builder.Header["kid"] = kid

	token, err := builder.SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}

// FromToken parses the token into their matching claims or returns an error if the
// the signature is invalid.
func FromToken(provider TokenKeyProvider, token string) (*Claims, error) {
	if provider == nil {
		return nil, fmt.Errorf("No token key provider configured")
	}

	claims := new(Claims)
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		var ok bool

		alg := ""
		if a := token.Header["alg"]; a != nil {
			alg, ok = a.(string)
			if !ok {
				return nil, errors.New("Invalid token alg")
			}
		}

		kid := ""
		if k := token.Header["kid"]; k != nil {
			kid, ok = k.(string)
			if !ok {
				return nil, errors.New("Invalid token kid")
			}
		}

		key, err := provider.Get(claims.Issuer, kid)
		if err != nil {
			return nil, err
		}

		switch k := key.(type) {
		case *rsa.PublicKey:
			if alg != jwt.SigningMethodRS512.Name {
				return nil, fmt.Errorf("Expected alg to be `%s` but got `%s`", jwt.SigningMethodRS512.Name, alg)
			}
			return k, nil
		case *ecdsa.PublicKey:
			if alg != jwt.SigningMethodES512.Name {
				return nil, fmt.Errorf("Expected alg to be `%s` but got `%s`", jwt.SigningMethodES512.Name, alg)
			}
			return k, nil
		default:
			return nil, ErrUnsupportedSigningMethod
		}
	})

	if err != nil {
		return nil, err
	}

	if err := claims.Valid(); err != nil {
		return nil, err
	}

	return claims, nil
}
