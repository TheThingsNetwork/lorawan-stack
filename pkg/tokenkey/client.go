// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"context"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

var _ claims.TokenKeyInfoProvider = new(Client)

// Client is a gRPC client to make calls to the IsTokenKeyInfo service that
// is used by the network components in order to introspect access tokens and keys.
//
// It implements claims.TokenKeyInfoProvider.
type Client struct {
	mu      sync.RWMutex
	secret  string
	clients map[string]ttnpb.IsTokenKeyInfoClient
}

// New returns a Client instance.
func New(secret string) *Client {
	return &Client{
		secret:  secret,
		clients: make(map[string]ttnpb.IsTokenKeyInfoClient),
	}
}

// TokenInfo returns the information about an access token.
// It returns error if token is expired.
func (c *Client) TokenInfo(accessToken string) (*types.AccessData, error) {
	_, payload, err := auth.DecodeTokenOrKey(accessToken)
	if err != nil {
		return nil, err
	}

	client, err := c.client(payload.Issuer)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetTokenInfo(context.Background(), &ttnpb.GetTokenInfoRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, err
	}

	return &types.AccessData{
		AccessToken: accessToken,
		UserID:      resp.UserID.UserID,
		ClientID:    resp.ClientID.ClientID,
		ExpiresIn:   time.Duration(resp.ExpiresIn) * time.Second,
		Scope:       resp.Scope,
	}, nil
}

// KeyInfo returns the information and entity ID about an API key.
func (c *Client) KeyInfo(key string) (string, *ttnpb.APIKey, error) {
	_, payload, err := auth.DecodeTokenOrKey(key)
	if err != nil {
		return "", nil, err
	}

	client, err := c.client(payload.Issuer)
	if err != nil {
		return "", nil, err
	}

	resp, err := client.GetKeyInfo(context.Background(), &ttnpb.GetKeyInfoRequest{
		Key: key,
	})
	if err != nil {
		return "", nil, err
	}

	return resp.EntityID, resp.Key, nil
}

func (c *Client) client(issuer string) (ttnpb.IsTokenKeyInfoClient, error) {
	c.mu.Lock()
	client, ok := c.clients[issuer]
	if !ok {
		conn, err := grpc.Dial(issuer, grpc.WithPerRPCCredentials(rpcmetadata.MD{
			AuthType:  "Basic",
			AuthValue: c.secret,
		}))
		if err != nil {
			return nil, err
		}

		c.clients[issuer] = ttnpb.NewIsTokenKeyInfoClient(conn)

		client = c.clients[issuer]
	}
	c.mu.Unlock()

	return client, nil
}
