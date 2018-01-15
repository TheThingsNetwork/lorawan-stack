// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

// ListApplicationRights returns the rights the caller user has to an application.
func (c *Client) ListApplicationRights(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationRightsResponse, error) {
	rights, err := c.cache.GetOrFetch(c.cache.ApplicationKey(req.ApplicationID), func() ([]ttnpb.Right, error) {
		resp, err := c.applications.ListApplicationRights(ctx, req, grpc.PerRPCCredentials(c.md))
		if err != nil {
			return nil, err
		}

		return resp.Rights, nil
	})

	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationRightsResponse{
		Rights: rights,
	}, nil
}
