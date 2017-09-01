// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package client

import (
	"context"
	"crypto"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	jwt "github.com/dgrijalva/jwt-go"

	"google.golang.org/grpc"
)

// ProviderClient is a auth.TokenKeyProvider that fetches the keys over gRPC.
type Client struct {
	sync.Mutex
	issuers     []string
	connections map[string]*grpc.ClientConn
}

// NewSimpleClient returns a new Client that can fetch token keys for the listed issuers.
func NewSimpleClient(issuers []string) *Client {
	return &Client{
		issuers:     issuers,
		connections: make(map[string]*grpc.ClientConn, len(issuers)),
	}
}

// Get implements auth.TokenKeyProvider
func (c *Client) Get(iss string, kid string) (crypto.PublicKey, error) {
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

	var key crypto.PublicKey
	switch resp.GetAlgorithm() {
	case jwt.SigningMethodRS512.Name, jwt.SigningMethodES256.Name, jwt.SigningMethodHS384.Name:
		key, err = jwt.ParseRSAPublicKeyFromPEM([]byte(resp.GetPublicKey()))
	case jwt.SigningMethodES512.Name, jwt.SigningMethodRS256.Name, jwt.SigningMethodRS384.Name:
		key, err = jwt.ParseECPublicKeyFromPEM([]byte(resp.GetPublicKey()))
	}
	if err != nil {
		return nil, err
	}

	if k := resp.GetKID(); k != "" {
		key = auth.WithKID(key, k)
	}

	return key, nil
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
		return nil, auth.ErrUnknownIdentityServer
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
