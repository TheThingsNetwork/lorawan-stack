// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Client is an Identity Server gRPC client.
type Client struct {
	md *rpcmetadata.MD

	cache Cache

	applications ttnpb.IsApplicationClient
	gateways     ttnpb.IsGatewayClient
}

// Option is the type that defines a Client option.
type Option func(*Client)

// WithCache sets the given cache to the Client.
func WithCache(cache Cache) Option {
	return func(c *Client) {
		c.cache = cache
	}
}

// New creates a new client.
func New(conn *grpc.ClientConn, md *rpcmetadata.MD, opts ...Option) *Client {
	client := &Client{
		md:           md,
		applications: ttnpb.NewIsApplicationClient(conn),
		gateways:     ttnpb.NewIsGatewayClient(conn),
		cache:        new(noopCache),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}
