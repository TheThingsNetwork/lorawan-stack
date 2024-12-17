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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var errManagedGatewayConfigurationServerNotEnabled = errors.DefineNotFound(
	"managed_gateway_configuration_server_not_enabled", "managed gateway configuration server is not enabled",
)

type noopManagedGCSServer struct {
	Component
	ttnpb.UnsafeManagedGatewayConfigurationServiceServer
}

var _ ttnpb.ManagedGatewayConfigurationServiceServer = (*noopManagedGCSServer)(nil)

// Get implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (*noopManagedGCSServer) Get(
	_ context.Context,
	_ *ttnpb.GetGatewayRequest,
) (*ttnpb.ManagedGateway, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// ScanWiFiAccessPoints implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (*noopManagedGCSServer) ScanWiFiAccessPoints(
	_ context.Context,
	_ *ttnpb.GatewayIdentifiers,
) (*ttnpb.ManagedGatewayWiFiAccessPoints, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// StreamEvents implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (*noopManagedGCSServer) StreamEvents(
	_ *ttnpb.GatewayIdentifiers,
	_ ttnpb.ManagedGatewayConfigurationService_StreamEventsServer,
) error {
	return errManagedGatewayConfigurationServerNotEnabled.New()
}

// Update implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (*noopManagedGCSServer) Update(
	_ context.Context,
	_ *ttnpb.UpdateManagedGatewayRequest,
) (*ttnpb.ManagedGateway, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

type noopManagedGatewayWiFiProfileServer struct {
	ttnpb.UnsafeManagedGatewayWiFiProfileConfigurationServiceServer
}

var _ ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer = (*noopManagedGatewayWiFiProfileServer)(nil)

// Create implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (*noopManagedGatewayWiFiProfileServer) Create(
	_ context.Context,
	_ *ttnpb.CreateManagedGatewayWiFiProfileRequest,
) (*ttnpb.ManagedGatewayWiFiProfile, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// Delete implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (*noopManagedGatewayWiFiProfileServer) Delete(
	_ context.Context,
	_ *ttnpb.DeleteManagedGatewayWiFiProfileRequest,
) (*emptypb.Empty, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// Get implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (*noopManagedGatewayWiFiProfileServer) Get(
	_ context.Context,
	_ *ttnpb.GetManagedGatewayWiFiProfileRequest,
) (*ttnpb.ManagedGatewayWiFiProfile, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// List implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (*noopManagedGatewayWiFiProfileServer) List(
	_ context.Context,
	_ *ttnpb.ListManagedGatewayWiFiProfilesRequest,
) (*ttnpb.ManagedGatewayWiFiProfiles, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// Update implements ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer.
func (*noopManagedGatewayWiFiProfileServer) Update(
	_ context.Context,
	_ *ttnpb.UpdateManagedGatewayWiFiProfileRequest,
) (*ttnpb.ManagedGatewayWiFiProfile, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

type noopManagedGatewayEthernetProfileServer struct {
	ttnpb.UnsafeManagedGatewayEthernetProfileConfigurationServiceServer
}

var _ ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer = (*noopManagedGatewayEthernetProfileServer)(nil) //nolint:lll

// Create implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (*noopManagedGatewayEthernetProfileServer) Create(
	_ context.Context,
	_ *ttnpb.CreateManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// Delete implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (*noopManagedGatewayEthernetProfileServer) Delete(
	_ context.Context,
	_ *ttnpb.DeleteManagedGatewayEthernetProfileRequest,
) (*emptypb.Empty, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// Get implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (*noopManagedGatewayEthernetProfileServer) Get(
	_ context.Context,
	_ *ttnpb.GetManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// List implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (*noopManagedGatewayEthernetProfileServer) List(
	_ context.Context,
	_ *ttnpb.ListManagedGatewayEthernetProfilesRequest,
) (*ttnpb.ManagedGatewayEthernetProfiles, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}

// Update implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (*noopManagedGatewayEthernetProfileServer) Update(
	_ context.Context,
	_ *ttnpb.UpdateManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	return nil, errManagedGatewayConfigurationServerNotEnabled.New()
}
