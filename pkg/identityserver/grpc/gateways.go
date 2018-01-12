// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

func (c *Client) CreateGateway(req *ttnpb.CreateGatewayRequest, credentials string) error {
	_, err := c.gateways.CreateGateway(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) GetGateway(req *ttnpb.GatewayIdentifier, credentials string) (*ttnpb.Gateway, error) {
	resp, err := c.gateways.GetGateway(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListGateways(credentials string) (*ttnpb.ListGatewaysResponse, error) {
	resp, err := c.gateways.ListGateways(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateGateway(req *ttnpb.UpdateGatewayRequest, credentials string) error {
	_, err := c.gateways.UpdateGateway(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) DeleteGateway(req *ttnpb.GatewayIdentifier, credentials string) error {
	_, err := c.gateways.DeleteGateway(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) GenerateGatewayAPIKey(req *ttnpb.GenerateGatewayAPIKeyRequest, credentials string) (*ttnpb.APIKey, error) {
	resp, err := c.gateways.GenerateGatewayAPIKey(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListGatewayAPIKeys(credentials string) (*ttnpb.ListGatewayAPIKeysResponse, error) {
	resp, err := c.gateways.ListGatewayAPIKeys(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateGatewayAPIKey(req *ttnpb.UpdateGatewayAPIKeyRequest, credentials string) error {
	_, err := c.gateways.UpdateGatewayAPIKey(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) RemoveGatewayAPIKey(req *ttnpb.RemoveGatewayAPIKeyRequest, credentials string) error {
	_, err := c.gateways.RemoveGatewayAPIKey(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) SetGatewayCollaborator(req *ttnpb.GatewayCollaborator, credentials string) error {
	_, err := c.gateways.SetGatewayCollaborator(context.Background(), req, creds(c.md, credentials))
	return err
}

func (c *Client) ListGatewayCollaborators(req *ttnpb.GatewayIdentifier, credentials string) error {
	resp, err := c.gateways.ListGatewayAPIKeys(context.Background(), req, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListGatewayRights(req *ttnpb.GatewayIdentifier, credentials string) error {
	rights, err := c.cache.GetOrFetch(c.cache.GatewayKey(req.GatewayID), func() ([]ttnpb.Right, error) {
		resp, err := c.applications.ListGatewayRights(context.Background(), req, creds(c.md, credentials))
		if err != nil {
			return nil, err
		}

		return resp.Rights, nil
	})

	if err != nil {
		return nil, err
	}

	return &ttnpb.ListGatewayRightsResponse{
		Rights: rights,
	}, nil
}
