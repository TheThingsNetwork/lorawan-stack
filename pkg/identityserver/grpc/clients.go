// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

func (c *Client) CreateClient(req *ttnpb.CreateClientRequest, credentials string) error {
	_, err := c.clients.CreateClient(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) GetClient(req *ttnpb.ClientIdentifier, credentials string) (*ttnpb.Client, error) {
	resp, err := c.clients.GetClient(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListClients(credentials string) (*ttnpb.ListClientsResponse, error) {
	resp, err := c.clients.ListClients(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateClient(req *ttnpb.UpdateClientRequest, credentials string) error {
	_, err := c.clients.UpdateClient(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) DeleteClient(req *ttnpb.ClientIdentifier, credentials string) error {
	_, err := c.clients.DeleteClient(context.Background(), req, creds(c.md, credentials))
	return err
}
