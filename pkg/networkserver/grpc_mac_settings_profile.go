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

// Package networkserver implements the LoRaWAN Network Server.
package networkserver

import (
	"context"
	"strconv"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	errMACSettingsProfileAlreadyExists = errors.DefineAlreadyExists("mac_settings_profile_already_exists", "MAC settings profile already exists") // nolint: lll
	errMACSettingsProfileNotFound      = errors.DefineNotFound("mac_settings_profile_not_found", "MAC settings profile not found")                // nolint: lll
)

func setTotalHeader(ctx context.Context, total int64) {
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(total, 10))) // nolint: errcheck
}

// NsMACSettingsProfileRegistry implements the MAC settings profile registry grpc service.
type NsMACSettingsProfileRegistry struct {
	ttnpb.UnimplementedNsMACSettingsProfileRegistryServer

	registry MACSettingsProfileRegistry
}

// Create creates a new MAC settings profile.
func (m *NsMACSettingsProfileRegistry) Create(ctx context.Context, req *ttnpb.CreateMACSettingsProfileRequest,
) (*ttnpb.CreateMACSettingsProfileResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.MacSettingsProfile.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	paths := []string{"ids", "mac_settings"}
	profile, err := m.registry.Set(
		ctx,
		req.MacSettingsProfile.Ids,
		paths,
		func(_ context.Context, profile *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error) {
			if profile != nil {
				return nil, nil, errMACSettingsProfileAlreadyExists.New()
			}
			return req.MacSettingsProfile, paths, nil
		})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to create MAC settings profile")
		return nil, err
	}

	return &ttnpb.CreateMACSettingsProfileResponse{
		MacSettingsProfile: profile,
	}, nil
}

// Get returns the MAC settings profile that matches the given identifiers.
func (m *NsMACSettingsProfileRegistry) Get(ctx context.Context, req *ttnpb.GetMACSettingsProfileRequest,
) (*ttnpb.GetMACSettingsProfileResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.MacSettingsProfileIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	); err != nil {
		return nil, err
	}
	paths := []string{"ids", "mac_settings"}
	if req.FieldMask != nil {
		paths = req.FieldMask.GetPaths()
	}
	profile, err := m.registry.Get(ctx, req.MacSettingsProfileIds, paths)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get MAC settings profile")
		return nil, err
	}

	return &ttnpb.GetMACSettingsProfileResponse{
		MacSettingsProfile: profile,
	}, nil
}

// Update updates the MAC settings profile that matches the given identifiers.
func (m *NsMACSettingsProfileRegistry) Update(ctx context.Context, req *ttnpb.UpdateMACSettingsProfileRequest,
) (*ttnpb.UpdateMACSettingsProfileResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.MacSettingsProfileIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	paths := []string{"ids", "mac_settings"}
	if req.FieldMask != nil {
		paths = req.FieldMask.GetPaths()
	}
	profile, err := m.registry.Set(
		ctx,
		req.MacSettingsProfile.Ids,
		paths,
		func(_ context.Context, profile *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error) {
			if profile == nil {
				return nil, nil, errMACSettingsProfileNotFound.New()
			}
			return req.MacSettingsProfile, paths, nil
		})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to create MAC settings profile")
		return nil, err
	}

	return &ttnpb.UpdateMACSettingsProfileResponse{
		MacSettingsProfile: profile,
	}, nil
}

// Delete deletes the MAC settings profile that matches the given identifiers.
func (m *NsMACSettingsProfileRegistry) Delete(ctx context.Context, req *ttnpb.DeleteMACSettingsProfileRequest,
) (*ttnpb.DeleteMACSettingsProfileResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.MacSettingsProfileIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	paths := []string{"ids", "mac_settings"}
	_, err := m.registry.Set(
		ctx,
		req.MacSettingsProfileIds,
		paths,
		func(_ context.Context, profile *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error) {
			if profile == nil {
				return nil, nil, errMACSettingsProfileNotFound.New()
			}
			return nil, nil, nil
		})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete MAC settings profile")
		return nil, err
	}

	return &ttnpb.DeleteMACSettingsProfileResponse{}, nil
}

// List lists the MAC settings profiles.
func (m *NsMACSettingsProfileRegistry) List(ctx context.Context, req *ttnpb.ListMACSettingsProfilesRequest,
) (*ttnpb.ListMACSettingsProfilesResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	); err != nil {
		return nil, err
	}
	paths := []string{"ids", "mac_settings"}
	if req.FieldMask != nil {
		paths = req.FieldMask.GetPaths()
	}

	var total int64
	ctx = m.registry.WithPagination(ctx, req.Limit, req.Page, &total)

	profiles, err := m.registry.List(ctx, req.ApplicationIds, paths)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to list MAC settings profiles")
		return nil, err
	}

	setTotalHeader(ctx, total)
	return &ttnpb.ListMACSettingsProfilesResponse{
		MacSettingsProfiles: profiles,
	}, nil
}
