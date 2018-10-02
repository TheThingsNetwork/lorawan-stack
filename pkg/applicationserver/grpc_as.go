// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetLink implements ttnpb.AsServer.
func (as *ApplicationServer) GetLink(ctx context.Context, req *ttnpb.GetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	return as.linkRegistry.Get(ctx, req.ApplicationIdentifiers)
}

// SetLink implements ttnpb.AsServer.
func (as *ApplicationServer) SetLink(ctx context.Context, req *ttnpb.SetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	link, err := as.linkRegistry.Set(ctx, req.ApplicationIdentifiers, req.FieldMask.Paths, func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &req.ApplicationLink, req.FieldMask.Paths, nil
	})
	if err != nil {
		return nil, err
	}
	// TODO: Restart link only when `network_server_address` is in field mask (https://github.com/TheThingsIndustries/lorawan-stack/issues/1177)
	if err := as.cancelLink(ctx, req.ApplicationIdentifiers); err != nil && !errors.IsNotFound(err) {
		log.FromContext(ctx).WithError(err).Warn("Failed to cancel link")
	}
	as.startLinkTask(ctx, req.ApplicationIdentifiers, &req.ApplicationLink)
	return link, nil
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
	if err := as.cancelLink(ctx, *ids); err != nil && !errors.IsNotFound(err) {
		log.FromContext(ctx).WithError(err).Warn("Failed to cancel link")
	}
	return ttnpb.Empty, nil
}
