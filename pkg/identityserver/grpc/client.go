// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/errors/grpcerrors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Client is an Identity Server gRPC client.
type Client struct {
	md rpcmetadata.MD

	target string

	cache Cache

	mu           sync.RWMutex
	conn         *grpc.ClientConn
	users        ttnpb.IsUserClient
	applications ttnpb.IsApplicationClient
	gateways     ttnpb.IsGatewayClient
	clients      ttnpb.IsClientClient
	settings     ttnpb.IsSettingsClient
}

// New creates a new client.
func New(target string, md rpcmetadata.MD, opts ...Option) *Client {
	client := &Client{
		md:     md,
		target: target,
		cache:  new(noopCache),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Connect opens the connection.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return nil
	}
	conn, err := grpc.Dial(c.target, grpc.WithUnaryInterceptor(grpcerrors.UnaryClientInterceptor()))
	if err != nil {
		return err
	}
	c.conn = conn
	c.users = ttnpb.NewIsUserClient(c.conn)
	c.applications = ttnpb.NewIsApplicationClient(c.conn)
	c.gateways = ttnpb.NewIsGatewayClient(c.conn)
	c.clients = ttnpb.NewIsClientClient(c.conn)
	c.settings = ttnpb.NewIsSettingsClient(c.conn)
	return nil
}

// Close closes the connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	if err != nil {
		return err
	}
	c.users = nil
	c.applications = nil
	c.gateways = nil
	c.clients = nil
	c.settings = nil
	c.conn = nil
	return nil
}

func creds(md rpcmetadata.MD, value string) grpc.PerRPCCredentials {
	md.AuthType = "Bearer"
	md.AuthValue = value
	return grpc.PerRPCCredentials(&md)
}
