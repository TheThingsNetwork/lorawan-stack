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

	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Delete implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Delete(
	ctx context.Context,
	req *ttnpb.DeleteManagedGatewayEthernetProfileRequest,
) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Get implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Get(
	ctx context.Context,
	req *ttnpb.GetManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// List implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) List(
	ctx context.Context,
	req *ttnpb.ListManagedGatewayEthernetProfilesRequest,
) (*ttnpb.ManagedGatewayEthernetProfiles, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Update implements ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer.
func (w *managedGatewayEthernetProfileServer) Update(
	ctx context.Context,
	req *ttnpb.UpdateManagedGatewayEthernetProfileRequest,
) (*ttnpb.ManagedGatewayEthernetProfile, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
