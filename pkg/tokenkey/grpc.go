// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"context"
	"crypto"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	jwt "github.com/dgrijalva/jwt-go"

	"google.golang.org/grpc"
)

// Client is a auth.TokenKeyProvider that fetches the keys over gRPC and caches the results.
type Client struct {
	sync.Mutex
	issuers     []string
	connections map[string]*grpc.ClientConn
	cache       Cache
}

// Option is an option for the Client.
type Option = func(*Client)

// NewClient returns a new Client that can fetch token keys for the listed issuers.
func NewClient(issuers []string, opts ...Option) *Client {
	client := &Client{
		issuers:     issuers,
		connections: make(map[string]*grpc.ClientConn, len(issuers)),
		cache:       NilCache,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// WithCache uses the cache on the client.
func WithCache(cache Cache) Option {
	return func(client *Client) {
		client.cache = cache
	}
}

// TokenKey implements auth.TokenKeyProvider.
func (c *Client) TokenKey(iss string, kid string) (crypto.PublicKey, error) {
	// try cache
	pem, alg, err := c.cache.Get(iss, kid)
	if err != nil {
		return nil, err
	}

	if pem == "" {
		conn, err := c.connect(iss)
		if err != nil {
			return nil, err
		}

		client := ttnpb.NewTokenKeyProviderClient(conn)

		resp, err := client.GetTokenKey(context.Background(), &ttnpb.TokenKeyRequest{
			KID: kid,
		})
		if err != nil {
			return nil, err
		}

		pem = resp.GetPublicKey()
		alg = resp.GetAlgorithm()
	}

	var key crypto.PublicKey
	switch alg {
	case jwt.SigningMethodRS512.Name, jwt.SigningMethodES256.Name, jwt.SigningMethodHS384.Name:
		key, err = jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
	case jwt.SigningMethodES512.Name, jwt.SigningMethodRS256.Name, jwt.SigningMethodRS384.Name:
		key, err = jwt.ParseECPublicKeyFromPEM([]byte(pem))
	}

	if err != nil {
		return nil, err
	}

	// update the cache
	err = c.cache.Set(iss, kid, alg, pem)
	if err != nil {
		return nil, err
	}

	return WithKID(key, kid), nil
}

// connect returns a connection to the issuer, using an old one if it exists.
// Not safe for concurrent use.
func (c *Client) connect(iss string) (*grpc.ClientConn, error) {
	c.Lock()
	defer c.Unlock()

	conn := c.connections[iss]
	if conn != nil {
		return conn, nil
	}

	if !c.knows(iss) {
		return nil, ErrUnknownIdentityServer
	}

	conn, err := grpc.Dial(iss)
	if err != nil {
		return nil, err
	}

	c.connections[iss] = conn

	return conn, nil
}

// knows returns true if the passed issuer is known by the client and false otherwise.
func (c *Client) knows(iss string) bool {
	for _, known := range c.issuers {
		if iss == known {
			return true
		}
	}
	return false
}
