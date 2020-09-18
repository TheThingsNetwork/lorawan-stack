// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package applicationserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetLink implements ttnpb.AsServer.
func (as *ApplicationServer) GetLink(ctx context.Context, req *ttnpb.GetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	return as.linkRegistry.Get(ctx, req.ApplicationIdentifiers, req.FieldMask.Paths)
}

// SetLink implements ttnpb.AsServer.
func (as *ApplicationServer) SetLink(ctx context.Context, req *ttnpb.SetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	return as.linkRegistry.Set(ctx, req.ApplicationIdentifiers, ttnpb.ApplicationLinkFieldPathsTopLevel,
		func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
			if link != nil {
				return &req.ApplicationLink, req.FieldMask.Paths, nil
			}
			return &req.ApplicationLink, append(req.FieldMask.Paths,
				"application_ids",
			), nil
		},
	)
}

// DeleteLink implements ttnpb.AsServer.
func (as *ApplicationServer) DeleteLink(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*types.Empty, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	_, err := as.linkRegistry.Set(ctx, *ids, nil, func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) { return nil, nil, nil })
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

var errNotImplemented = errors.DefineUnimplemented("linking_not_implemented", "linking is not implemented")

// GetLinkStats implements ttnpb.AsServer.
func (as *ApplicationServer) GetLinkStats(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.ApplicationLinkStats, error) {
	return nil, errNotImplemented.New()
}

// HandleUplink implements ttnpb.NsAsServer.
func (as *ApplicationServer) HandleUplink(ctx context.Context, req *ttnpb.NsAsHandleUplinkRequest) (*types.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	link, err := as.linkRegistry.Get(ctx, req.ApplicationUps[0].ApplicationIdentifiers, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return nil, err
	}
	// TODO: Merge downlink queue invalidations (https://github.com/TheThingsNetwork/lorawan-stack/issues/1523)
	for _, up := range req.ApplicationUps {
		if err := as.processUp(ctx, up, link); err != nil {
			return nil, err
		}
	}
	return ttnpb.Empty, nil
}
