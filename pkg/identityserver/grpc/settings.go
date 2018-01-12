// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

func (c *Client) GetSettings(credentials string) (*ttnpb.IdentityServerSettings, error) {
	resp, err := c.settings.GetSettings(context.Background(), &pbtypes.Empty{}, creds(c.md, credentials))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateSettings(req *ttnpb.UpdateSettingsRequest, credentials string) error {
	_, err := c.settings.UpdateSettings(context.Background(), req, creds(c.md, credentials))
	return err
}
