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

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// VendorIDProfileFetcher handles the validation and fetching of end device profile in the Device Repository.
type VendorIDProfileFetcher struct {
	Component        Component
	DeviceRepository ttnpb.DeviceRepositoryClient
}

// ShouldFetchProfile dictactes if the end device has the necessary fields to fetch its profile.
func (*VendorIDProfileFetcher) ShouldFetchProfile(device *ttnpb.EndDevice) bool {
	return device.GetVersionIds().GetVendorId() != 0 && device.GetVersionIds().GetVendorProfileId() != 0
}

// FetchProfile provides the end device profile.
func (pf *VendorIDProfileFetcher) FetchProfile(
	ctx context.Context,
	device *ttnpb.EndDevice,
) (*ttnpb.EndDeviceTemplate, error) {
	profileIdentifiers := &ttnpb.GetTemplateRequest_EndDeviceProfileIdentifiers{
		VendorId:        device.GetVersionIds().GetVendorId(),
		VendorProfileId: device.GetVersionIds().GetVendorProfileId(),
	}
	if pf.DeviceRepository != nil {
		return pf.DeviceRepository.GetTemplate(ctx, &ttnpb.GetTemplateRequest{EndDeviceProfileIds: profileIdentifiers})
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
		GetTemplate(ctx, &ttnpb.GetTemplateRequest{EndDeviceProfileIds: profileIdentifiers}, opt)
}
