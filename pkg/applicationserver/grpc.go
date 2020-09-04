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
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
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
	// Get all the fields here for starting the link task.
	link, err := as.linkRegistry.Set(ctx, req.ApplicationIdentifiers, ttnpb.ApplicationLinkFieldPathsTopLevel,
		func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
			return &req.ApplicationLink, req.FieldMask.Paths, nil
		},
	)
	if err != nil {
		return nil, err
	}
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
	// TODO: Do we return a deprecated error ?
	return &ttnpb.ApplicationLinkStats{}, nil
}

func (as *ApplicationServer) HandleUplink(ctx context.Context, req *ttnpb.NsAsHandleUplinkRequest) (*types.Empty, error) {
	for _, up := range req.ApplicationUplinks {
		ctx := events.ContextWithCorrelationID(ctx, append(up.CorrelationIDs, fmt.Sprintf("as:up:%s", events.NewCorrelationID()))...)
		logger := log.FromContext(ctx)
		up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)
		// TODO: How can we use the caller name here ?
		registerReceiveUp(ctx, up, "replace-me")

		now := time.Now().UTC()
		up.ReceivedAt = &now

		pass, err := as.handleUp(ctx, up, nil)
		if err != nil {
			logger.WithError(err).Warn("Failed to process upstream message")
			registerDropUp(ctx, up, err)
			continue
		}
		if !pass {
			continue
		}

		if err := as.SendUp(ctx, up); err != nil {
			logger.WithError(err).Warn("Failed to send upstream message")
			registerDropUp(ctx, up, err)
			continue
		}
		registerForwardUp(ctx, up)
	}
	return ttnpb.Empty, nil
}
