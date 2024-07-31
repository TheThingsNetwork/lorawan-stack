// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package managed

import (
	"context"
	"strconv"

	northboundv1 "go.thethings.industries/pkg/api/gen/tti/gateway/controller/northbound/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type managedGatewayEthernetProfileServer struct {
	ttnpb.UnsafeManagedGatewayEthernetProfileConfigurationServiceServer
	client *ttgc.Client
}

var _ ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer = (*managedGatewayEthernetProfileServer)(nil)

// Create implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Create(
	ctx context.Context,
	req *ttnpb.CreateManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	if req.Profile.ProfileName == "" {
		return nil, errNoProfileName.New()
	}
	profile := req.Profile
	res, err := northboundv1.NewEthernetProfileServiceClient(w.client).Create(
		ctx,
		&northboundv1.EthernetProfileServiceCreateRequest{
			Domain:          w.client.Domain(ctx),
			Group:           group(req.Collaborator),
			EthernetProfile: fromEthernetProfile(profile),
		},
	)
	if err != nil {
		return nil, err
	}
	profile.ProfileId = toProfileID(res.ProfileId)
	return profile, nil
}

// Delete implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Delete(
	ctx context.Context,
	req *ttnpb.DeleteManagedGatewayEthernetProfileRequest,
) (*emptypb.Empty, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	_, err := northboundv1.NewEthernetProfileServiceClient(w.client).Delete(
		ctx,
		&northboundv1.EthernetProfileServiceDeleteRequest{
			Domain:    w.client.Domain(ctx),
			Group:     group(req.Collaborator),
			ProfileId: fromProfileID(req.ProfileId),
		},
	)
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

// Get implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Get(
	ctx context.Context,
	req *ttnpb.GetManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	profileID := fromProfileID(req.ProfileId)
	res, err := northboundv1.NewEthernetProfileServiceClient(w.client).Get(
		ctx,
		&northboundv1.EthernetProfileServiceGetRequest{
			Domain:    w.client.Domain(ctx),
			Group:     group(req.Collaborator),
			ProfileId: profileID,
		},
	)
	if err != nil {
		return nil, err
	}
	profile := toEthernetProfile(profileID, res.EthernetProfile)
	return profile, nil
}

// List implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) List(
	ctx context.Context,
	req *ttnpb.ListManagedGatewayEthernetProfilesRequest,
) (*ttnpb.ManagedGatewayEthernetProfiles, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	page := req.Page
	if page == 0 {
		page = 1
	}
	res, err := northboundv1.NewEthernetProfileServiceClient(w.client).List(
		ctx,
		&northboundv1.EthernetProfileServiceListRequest{
			Domain: w.client.Domain(ctx),
			Group:  group(req.Collaborator),
			Limit:  req.Limit,
			Offset: (page - 1) * req.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	profiles := make([]*ttnpb.ManagedGatewayEthernetProfile, 0, len(res.Entries))
	for _, entry := range res.Entries {
		profile := toEthernetProfile(entry.ProfileId, entry.EthernetProfile)
		profiles = append(profiles, profile)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(int64(res.Total), 10))) //nolint:errcheck
	return &ttnpb.ManagedGatewayEthernetProfiles{
		Profiles: profiles,
	}, nil
}

// Update implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Update(
	ctx context.Context,
	req *ttnpb.UpdateManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	client := northboundv1.NewEthernetProfileServiceClient(w.client)
	profileID := fromProfileID(req.Profile.ProfileId)
	getRes, err := client.Get(ctx, &northboundv1.EthernetProfileServiceGetRequest{
		Domain:    w.client.Domain(ctx),
		Group:     group(req.Collaborator),
		ProfileId: profileID,
	})
	if err != nil {
		return nil, err
	}
	profile := toEthernetProfile(profileID, getRes.EthernetProfile)
	if err := profile.SetFields(req.Profile, req.FieldMask.GetPaths()...); err != nil {
		return nil, err
	}
	_, err = client.Update(ctx, &northboundv1.EthernetProfileServiceUpdateRequest{
		Domain:          w.client.Domain(ctx),
		Group:           group(req.Collaborator),
		ProfileId:       profileID,
		EthernetProfile: fromEthernetProfile(profile),
	})
	if err != nil {
		return nil, err
	}
	return profile, nil
}
