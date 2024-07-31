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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type managedGatewayWiFiProfileServer struct {
	ttnpb.UnsafeManagedGatewayWiFiProfileConfigurationServiceServer
	client *ttgc.Client
}

var _ ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer = (*managedGatewayWiFiProfileServer)(nil)

var errNoSSID = errors.DefineInvalidArgument("no_ssid", "no SSID set")

// Create implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (w *managedGatewayWiFiProfileServer) Create(
	ctx context.Context,
	req *ttnpb.CreateManagedGatewayWiFiProfileRequest,
) (*ttnpb.ManagedGatewayWiFiProfile, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	if req.Profile.ProfileName == "" {
		return nil, errNoProfileName.New()
	}
	if req.Profile.Ssid == "" {
		return nil, errNoSSID.New()
	}
	profile := req.Profile
	res, err := northboundv1.NewWifiProfileServiceClient(w.client).Create(
		ctx,
		&northboundv1.WifiProfileServiceCreateRequest{
			Domain:      w.client.Domain(ctx),
			Group:       group(req.Collaborator),
			WifiProfile: fromWiFiProfile(profile),
		},
	)
	if err != nil {
		return nil, err
	}
	profile.ProfileId = toProfileID(res.ProfileId)
	profile.Password = ""
	return profile, nil
}

// Delete implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (w *managedGatewayWiFiProfileServer) Delete(
	ctx context.Context,
	req *ttnpb.DeleteManagedGatewayWiFiProfileRequest,
) (*emptypb.Empty, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	_, err := northboundv1.NewWifiProfileServiceClient(w.client).Delete(
		ctx,
		&northboundv1.WifiProfileServiceDeleteRequest{
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

// Get implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (w *managedGatewayWiFiProfileServer) Get(
	ctx context.Context,
	req *ttnpb.GetManagedGatewayWiFiProfileRequest,
) (*ttnpb.ManagedGatewayWiFiProfile, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	profileID := fromProfileID(req.ProfileId)
	res, err := northboundv1.NewWifiProfileServiceClient(w.client).Get(
		ctx,
		&northboundv1.WifiProfileServiceGetRequest{
			Domain:    w.client.Domain(ctx),
			Group:     group(req.Collaborator),
			ProfileId: profileID,
		},
	)
	if err != nil {
		return nil, err
	}
	profile := toWiFiProfile(profileID, res.WifiProfile)
	profile.Password = ""
	return profile, nil
}

// List implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (w *managedGatewayWiFiProfileServer) List(
	ctx context.Context,
	req *ttnpb.ListManagedGatewayWiFiProfilesRequest,
) (*ttnpb.ManagedGatewayWiFiProfiles, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	page := req.Page
	if page == 0 {
		page = 1
	}
	res, err := northboundv1.NewWifiProfileServiceClient(w.client).List(
		ctx,
		&northboundv1.WifiProfileServiceListRequest{
			Domain: w.client.Domain(ctx),
			Group:  group(req.Collaborator),
			Limit:  req.Limit,
			Offset: (page - 1) * req.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	profiles := make([]*ttnpb.ManagedGatewayWiFiProfile, 0, len(res.Entries))
	for _, entry := range res.Entries {
		profile := toWiFiProfile(entry.ProfileId, entry.WifiProfile)
		profile.Password = ""
		profiles = append(profiles, profile)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(int64(res.Total), 10))) //nolint:errcheck
	return &ttnpb.ManagedGatewayWiFiProfiles{
		Profiles: profiles,
	}, nil
}

// Update implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (w *managedGatewayWiFiProfileServer) Update(
	ctx context.Context,
	req *ttnpb.UpdateManagedGatewayWiFiProfileRequest,
) (*ttnpb.ManagedGatewayWiFiProfile, error) {
	if err := requireProfileRights(ctx, req.Collaborator); err != nil {
		return nil, err
	}
	client := northboundv1.NewWifiProfileServiceClient(w.client)
	profileID := fromProfileID(req.Profile.ProfileId)
	getRes, err := client.Get(ctx, &northboundv1.WifiProfileServiceGetRequest{
		Domain:    w.client.Domain(ctx),
		Group:     group(req.Collaborator),
		ProfileId: profileID,
	})
	if err != nil {
		return nil, err
	}
	profile := toWiFiProfile(profileID, getRes.WifiProfile)
	if err := profile.SetFields(req.Profile, req.FieldMask.GetPaths()...); err != nil {
		return nil, err
	}
	_, err = client.Update(ctx, &northboundv1.WifiProfileServiceUpdateRequest{
		Domain:      w.client.Domain(ctx),
		Group:       group(req.Collaborator),
		ProfileId:   profileID,
		WifiProfile: fromWiFiProfile(profile),
	})
	if err != nil {
		return nil, err
	}
	profile.Password = ""
	return profile, nil
}
