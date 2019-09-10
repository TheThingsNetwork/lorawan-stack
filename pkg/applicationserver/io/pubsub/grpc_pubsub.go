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

package pubsub

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// appendImplicitPubSubGetPaths appends implicit ttnpb.ApplicationPubSub get paths to paths.
func appendImplicitPubSubGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 3+len(paths)),
		"format",
		"provider",
		"base_topic",
	), paths...)
}

// GetFormats implements ttnpb.ApplicationPubSubRegistryServer.
func (ps *PubSub) GetFormats(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ApplicationPubSubFormats, error) {
	fs := make(map[string]string, len(formats))
	for key, val := range formats {
		fs[key] = val.Name
	}
	return &ttnpb.ApplicationPubSubFormats{
		Formats: fs,
	}, nil
}

// Get implements ttnpb.ApplicationPubSubRegistryServer.
func (ps *PubSub) Get(ctx context.Context, req *ttnpb.GetApplicationPubSubRequest) (*ttnpb.ApplicationPubSub, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	return ps.registry.Get(ctx, req.ApplicationPubSubIdentifiers, appendImplicitPubSubGetPaths(req.FieldMask.Paths...))
}

// List implements ttnpb.ApplicationPubSubRegistryServer.
func (ps *PubSub) List(ctx context.Context, req *ttnpb.ListApplicationPubSubsRequest) (*ttnpb.ApplicationPubSubs, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	pubsubs, err := ps.registry.List(ctx, req.ApplicationIdentifiers, appendImplicitPubSubGetPaths(req.FieldMask.Paths...))
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationPubSubs{
		Pubsubs: pubsubs,
	}, nil
}

// Set implements ttnpb.ApplicationPubSubRegistryServer.
func (ps *PubSub) Set(ctx context.Context, req *ttnpb.SetApplicationPubSubRequest) (*ttnpb.ApplicationPubSub, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers,
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	); err != nil {
		return nil, err
	}
	// Get all the fields here for starting the integration task.
	pubsub, err := ps.registry.Set(ctx, req.ApplicationPubSubIdentifiers, appendImplicitPubSubGetPaths(req.FieldMask.Paths...),
		func(pubsub *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
			if pubsub != nil {
				return &req.ApplicationPubSub, req.FieldMask.Paths, nil
			}
			return &req.ApplicationPubSub, append(req.FieldMask.Paths,
				"ids.application_ids",
				"ids.pub_sub_id",
			), nil
		},
	)
	if err != nil {
		return nil, err
	}
	if err := ps.stop(ctx, req.ApplicationPubSubIdentifiers); err != nil && !errors.IsNotFound(err) {
		log.FromContext(ctx).WithFields(log.Fields(
			"application_uid", unique.ID(ctx, req.ApplicationIdentifiers),
			"pub_sub_id", req.PubSubID,
		)).WithError(err).Warn("Failed to cancel integration")
	}
	ps.startTask(ps.ctx, req.ApplicationPubSubIdentifiers)
	events.Publish(evtSetPubSub(ctx, req.ApplicationPubSubIdentifiers, nil))
	return pubsub, nil
}

// Delete implements ttnpb.ApplicationPubSubRegistryServer.
func (ps *PubSub) Delete(ctx context.Context, ids *ttnpb.ApplicationPubSubIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers,
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	); err != nil {
		return nil, err
	}
	if err := ps.stop(ctx, *ids); err != nil {
		log.FromContext(ctx).WithFields(log.Fields(
			"application_uid", unique.ID(ctx, ids.ApplicationIdentifiers),
			"pub_sub_id", ids.PubSubID,
		)).WithError(err).Warn("Failed to cancel integration")
	}
	_, err := ps.registry.Set(ctx, *ids, nil,
		func(pubsub *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
			return nil, nil, nil
		},
	)
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeletePubSub(ctx, *ids, nil))
	return ttnpb.Empty, nil
}
