// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package deviceclaimingserver_test

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockEndDeviceClaimer is a mock Claimer.
type MockEndDeviceClaimer struct {
	JoinEUIs []types.EUI64

	ClaimFunc        func(context.Context, types.EUI64, types.EUI64, string) error
	BatchUnclaimFunc func(
		context.Context,
		[]*ttnpb.EndDeviceIdentifiers,
	) error
}

// SupportsJoinEUI returns whether the Join Server supports this JoinEUI.
func (m MockEndDeviceClaimer) SupportsJoinEUI(joinEUI types.EUI64) bool {
	for _, eui := range m.JoinEUIs {
		if eui.Equal(joinEUI) {
			return true
		}
	}
	return false
}

// Claim claims an End Device.
func (m MockEndDeviceClaimer) Claim(
	ctx context.Context, joinEUI, devEUI types.EUI64, claimAuthenticationCode string,
) error {
	return m.ClaimFunc(ctx, joinEUI, devEUI, claimAuthenticationCode)
}

// GetClaimStatus returns the claim status an End Device.
func (MockEndDeviceClaimer) GetClaimStatus(_ context.Context,
	ids *ttnpb.EndDeviceIdentifiers,
) (*ttnpb.GetClaimStatusResponse, error) {
	return &ttnpb.GetClaimStatusResponse{
		EndDeviceIds: ids,
	}, nil
}

// Unclaim releases the claim on an End Device.
func (MockEndDeviceClaimer) Unclaim(_ context.Context,
	_ *ttnpb.EndDeviceIdentifiers,
) (err error) {
	return nil
}

// Unclaim releases the claim on an End Device.
func (m MockEndDeviceClaimer) BatchUnclaim(
	ctx context.Context,
	ids []*ttnpb.EndDeviceIdentifiers,
) error {
	return m.BatchUnclaimFunc(ctx, ids)
}

// MockGatewayClaimer is a mock gateway Claimer.
type MockGatewayClaimer struct {
	EUIs []types.EUI64

	ClaimFunc   func(context.Context, types.EUI64, string, string) error
	UnclaimFunc func(context.Context, types.EUI64, string) error
}

// Claim implements gateways.Claimer.
func (claimer MockGatewayClaimer) Claim(
	ctx context.Context,
	eui types.EUI64,
	ownerToken string,
	clusterAddress string,
) error {
	return claimer.ClaimFunc(ctx, eui, ownerToken, clusterAddress)
}

// Unclaim implements gateways.Claimer.
func (claimer MockGatewayClaimer) Unclaim(ctx context.Context, eui types.EUI64, clusterAddress string) error {
	return claimer.UnclaimFunc(ctx, eui, clusterAddress)
}

type mockGatewayRegistry struct {
	gateways     []*ttnpb.Gateway
	authorizedMD rpcmetadata.MD

	createFunc func(ctx context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error)
	deleteFunc func(ctx context.Context, in *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error)
	getFunc    func(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
}

var (
	errInvalidCredentials = errors.DefineUnauthenticated("invalid_credentials", "invalid credentials")
	errGatewayNotFound    = errors.DefineNotFound("gateway_not_found", "gateway not found")
	errClaim              = errors.DefineAborted("claim", "claim")
	errCreate             = errors.DefineAborted("create_gateway", "create gateway")
	errUnclaim            = errors.DefineAborted("unclaim", "unclaim gateway")
)

// AssertGatewayRights implements GatewayRegistry.
func (mock mockGatewayRegistry) AssertGatewayRights(
	ctx context.Context,
	_ *ttnpb.GatewayIdentifiers,
	_ ...ttnpb.Right,
) error {
	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType == mock.authorizedMD.AuthType && md.AuthValue == mock.authorizedMD.AuthValue {
		return nil
	}
	return errInvalidCredentials.New()
}

// GetIdentifiersForEUI implements GatewayRegistry.
func (mock mockGatewayRegistry) GetIdentifiersForEUI(
	_ context.Context,
	eui types.EUI64,
) (*ttnpb.GatewayIdentifiers, error) {
	for _, gateway := range mock.gateways {
		if types.MustEUI64(gateway.GetIds().Eui).Equal(eui) {
			return gateway.Ids, nil
		}
	}
	return nil, errGatewayNotFound.New()
}

// Create implements GatewayRegistry.
func (mock mockGatewayRegistry) Create(ctx context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
	return mock.createFunc(ctx, in)
}

// Delete implements GatewayRegistry.
func (mock mockGatewayRegistry) Delete(ctx context.Context, in *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
	return mock.deleteFunc(ctx, in)
}

// Get implements GatewayRegistry.
func (mock mockGatewayRegistry) Get(ctx context.Context, in *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	return mock.getFunc(ctx, in)
}
