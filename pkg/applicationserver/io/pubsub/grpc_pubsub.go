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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// appendImplicitPubSubGetPaths appends implicit ttnpb.ApplicationPubSub get paths to paths.
func appendImplicitPubSubGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 2+len(paths)),
		"format",
		"provider",
	), paths...)
}

type pubsubRegistryRPC struct {
	pubsubs Registry
}

// NewPubSubRegistryRPC returns a new PubSub registry gRPC server.
func NewPubSubRegistryRPC(pubsubs Registry) ttnpb.ApplicationPubSubRegistryServer {
	return &pubsubRegistryRPC{
		pubsubs: pubsubs,
	}
}

func (s pubsubRegistryRPC) GetFormats(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ApplicationPubSubFormats, error) {
	fs := make(map[string]string, len(formats))
	for key, val := range formats {
		fs[key] = val.Name
	}
	return &ttnpb.ApplicationPubSubFormats{
		Formats: fs,
	}, nil
}

func (s pubsubRegistryRPC) Get(ctx context.Context, req *ttnpb.GetApplicationPubSubRequest) (*ttnpb.ApplicationPubSub, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	return s.pubsubs.Get(ctx, req.ApplicationPubSubIdentifiers, appendImplicitPubSubGetPaths(req.FieldMask.Paths...))
}

func (s pubsubRegistryRPC) List(ctx context.Context, req *ttnpb.ListApplicationPubSubsRequest) (*ttnpb.ApplicationPubSubs, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	pubsubs, err := s.pubsubs.List(ctx, req.ApplicationIdentifiers, appendImplicitPubSubGetPaths(req.FieldMask.Paths...))
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationPubSubs{
		Pubsubs: pubsubs,
	}, nil
}

func (s pubsubRegistryRPC) Set(ctx context.Context, req *ttnpb.SetApplicationPubSubRequest) (*ttnpb.ApplicationPubSub, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	return s.pubsubs.Set(ctx, req.ApplicationPubSubIdentifiers, appendImplicitPubSubGetPaths(req.FieldMask.Paths...),
		func(pubsub *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
			if pubsub != nil {
				return &req.ApplicationPubSub, req.FieldMask.Paths, nil
			}
			return &req.ApplicationPubSub, append(req.FieldMask.Paths,
				"ids.application_ids",
				"ids.pubsub_id",
			), nil
		},
	)
}

func (s pubsubRegistryRPC) Delete(ctx context.Context, req *ttnpb.ApplicationPubSubIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	_, err := s.pubsubs.Set(ctx, *req, nil,
		func(pubsub *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
			return nil, nil, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
