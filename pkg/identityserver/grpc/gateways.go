// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

func (c *Client) ListGatewayRights(ctx context.Context, req *ttnpb.GatewayIdentifier) error {
	rights, err := c.cache.GetOrFetch(c.cache.GatewayKey(req.GatewayID), func() ([]ttnpb.Right, error) {
		resp, err := c.applications.ListGatewayRights(ctx, req, grpc.PerRPCCredentials(c.md))
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
