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
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	errLinkDeleted = errors.DefineAborted("link_deleted", "link deleted")
	errLinkReset   = errors.DefineUnavailable("link_reset", "link reset")
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
	// Get all the fields here for starting the link task.
	link, err := as.linkRegistry.Set(ctx, req.ApplicationIdentifiers, ttnpb.ApplicationLinkFieldPathsTopLevel,
		func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
			return &req.ApplicationLink, req.FieldMask.Paths, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if err := as.cancelLink(ctx, req.ApplicationIdentifiers, errLinkReset); err != nil && !errors.IsNotFound(err) {
		log.FromContext(ctx).WithError(err).Warn("Failed to cancel link")
	}
	as.startLinkTask(as.Context(), req.ApplicationIdentifiers)

	res := &ttnpb.ApplicationLink{}
	if err := res.SetFields(link, req.FieldMask.Paths...); err != nil {
		return nil, err
	}
	return res, nil
}

// DeleteLink implements ttnpb.AsServer.
func (as *ApplicationServer) DeleteLink(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*types.Empty, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	if err := as.cancelLink(ctx, *ids, errLinkDeleted); err != nil && !errors.IsNotFound(err) {
		log.FromContext(ctx).WithError(err).Warn("Failed to cancel link")
	}
	_, err := as.linkRegistry.Set(ctx, *ids, nil, func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) { return nil, nil, nil })
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

// GetLinkStats implements ttnpb.AsServer.
func (as *ApplicationServer) GetLinkStats(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.ApplicationLinkStats, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}

	link, err := as.getLink(ctx, *ids)
	if err != nil {
		return nil, err
	}
	<-link.connReady

	stats := &ttnpb.ApplicationLinkStats{}
	lt := link.GetLinkTime()
	stats.LinkedAt = &lt
	stats.NetworkServerAddress = link.NetworkServerAddress
	if n, t, ok := link.GetUpStats(); ok {
		stats.UpCount = n
		stats.LastUpReceivedAt = &t
	}
	if n, t, ok := link.GetDownlinkStats(); ok {
		stats.DownlinkCount = n
		stats.LastDownlinkForwardedAt = &t
	}
	return stats, nil
}
