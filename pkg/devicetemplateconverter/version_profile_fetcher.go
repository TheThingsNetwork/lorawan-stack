// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package devicetemplateconverter

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Component abstracts the underlying *component.Component.
type Component interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

// VersionIDProfileFetcher handles the validation and fetching of end device profile in the Device Repository.
type VersionIDProfileFetcher struct {
	Component        Component
	DeviceRepository ttnpb.DeviceRepositoryClient
}

// ShouldFetchProfile dictactes if the end device has the necessary fields to fetch its profile.
func (*VersionIDProfileFetcher) ShouldFetchProfile(device *ttnpb.EndDevice) bool {
	return device.GetVersionIds().GetBrandId() != "" &&
		device.GetVersionIds().GetModelId() != "" &&
		device.GetVersionIds().GetFirmwareVersion() != "" &&
		device.GetVersionIds().GetBandId() != ""
}

// FetchProfile provides the end device profile.
func (pf *VersionIDProfileFetcher) FetchProfile(ctx context.Context, device *ttnpb.EndDevice) (*ttnpb.EndDeviceTemplate, error) {
	versionIDs := &ttnpb.EndDeviceVersionIdentifiers{
		BrandId:         device.GetVersionIds().GetBrandId(),
		ModelId:         device.GetVersionIds().GetModelId(),
		HardwareVersion: device.GetVersionIds().GetHardwareVersion(),
		FirmwareVersion: device.GetVersionIds().GetFirmwareVersion(),
		BandId:          device.GetVersionIds().GetBandId(),
	}
	if pf.DeviceRepository != nil {
		return pf.DeviceRepository.GetTemplate(ctx, &ttnpb.GetTemplateRequest{VersionIds: versionIDs})
	}

	conn, err := pf.Component.GetPeerConn(ctx, ttnpb.ClusterRole_DEVICE_REPOSITORY, nil)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get Device Repository peer")
		return nil, err
	}

	opt, err := rpcmetadata.WithForwardedAuth(ctx, pf.Component.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}

	return ttnpb.NewDeviceRepositoryClient(conn).
		GetTemplate(ctx, &ttnpb.GetTemplateRequest{VersionIds: versionIDs}, opt)
}
