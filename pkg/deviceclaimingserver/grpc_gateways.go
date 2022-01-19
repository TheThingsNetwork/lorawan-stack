// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package deviceclaimingserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// noopGCLS is a no-op GCLS.
type noopGCLS struct {
}

// Claim implements GatewayClaimingServer.
func (noopGCLS) Claim(ctx context.Context, req *ttnpb.ClaimGatewayRequest) (ids *ttnpb.GatewayIdentifiers, retErr error) {
	return nil, errMethodUnavailable.New()
}

// AuthorizeGateway implements GatewayClaimingServer.
func (noopGCLS) AuthorizeGateway(ctx context.Context, req *ttnpb.AuthorizeGatewayRequest) (*pbtypes.Empty, error) {
	return nil, errMethodUnavailable.New()
}

// UnauthorizeGateway implements GatewayClaimingServer.
func (noopGCLS) UnauthorizeGateway(ctx context.Context, gtwIDs *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return nil, errMethodUnavailable.New()
}

// gatewayClaimingServer is the front facing entity for gRPC requests.
type gatewayClaimingServer struct {
	DCS *DeviceClaimingServer
}

// Claim implements GatewayClaimingServer.
func (gcls gatewayClaimingServer) Claim(ctx context.Context, req *ttnpb.ClaimGatewayRequest) (ids *ttnpb.GatewayIdentifiers, retErr error) {
	return gcls.DCS.gatewayClaimingServerUpstream.Claim(ctx, req)
}

// AuthorizeGateway implements GatewayClaimingServer.
func (gcls gatewayClaimingServer) AuthorizeGateway(ctx context.Context, req *ttnpb.AuthorizeGatewayRequest) (*pbtypes.Empty, error) {
	return gcls.DCS.gatewayClaimingServerUpstream.AuthorizeGateway(ctx, req)
}

// UnauthorizeGateway implements GatewayClaimingServer.
func (gcls gatewayClaimingServer) UnauthorizeGateway(ctx context.Context, gtwIDs *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return gcls.DCS.gatewayClaimingServerUpstream.UnauthorizeGateway(ctx, gtwIDs)
}
