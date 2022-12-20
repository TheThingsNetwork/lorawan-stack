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

package packetbrokeragent

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type disabledServer struct {
	ttnpb.UnimplementedPbaServer
	ttnpb.UnimplementedGsPbaServer
	ttnpb.UnimplementedNsPbaServer
}

var errNotEnabled = errors.DefineFailedPrecondition("not_enabled", "Packet Broker is not enabled")

func (s disabledServer) GetInfo(context.Context, *pbtypes.Empty) (*ttnpb.PacketBrokerInfo, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) Register(context.Context, *ttnpb.PacketBrokerRegisterRequest) (*ttnpb.PacketBrokerNetwork, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) Deregister(context.Context, *pbtypes.Empty) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) GetHomeNetworkDefaultRoutingPolicy(context.Context, *pbtypes.Empty) (*ttnpb.PacketBrokerDefaultRoutingPolicy, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) SetHomeNetworkDefaultRoutingPolicy(context.Context, *ttnpb.SetPacketBrokerDefaultRoutingPolicyRequest) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) DeleteHomeNetworkDefaultRoutingPolicy(context.Context, *pbtypes.Empty) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) ListHomeNetworkRoutingPolicies(context.Context, *ttnpb.ListHomeNetworkRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) GetHomeNetworkRoutingPolicy(context.Context, *ttnpb.PacketBrokerNetworkIdentifier) (*ttnpb.PacketBrokerRoutingPolicy, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) SetHomeNetworkRoutingPolicy(context.Context, *ttnpb.SetPacketBrokerRoutingPolicyRequest) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) DeleteHomeNetworkRoutingPolicy(context.Context, *ttnpb.PacketBrokerNetworkIdentifier) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s *disabledServer) GetHomeNetworkDefaultGatewayVisibility(context.Context, *pbtypes.Empty) (*ttnpb.PacketBrokerDefaultGatewayVisibility, error) {
	return nil, errNotEnabled.New()
}

func (s *disabledServer) SetHomeNetworkDefaultGatewayVisibility(context.Context, *ttnpb.SetPacketBrokerDefaultGatewayVisibilityRequest) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s *disabledServer) DeleteHomeNetworkDefaultGatewayVisibility(context.Context, *pbtypes.Empty) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) ListNetworks(context.Context, *ttnpb.ListPacketBrokerNetworksRequest) (*ttnpb.PacketBrokerNetworks, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) ListHomeNetworks(context.Context, *ttnpb.ListPacketBrokerHomeNetworksRequest) (*ttnpb.PacketBrokerNetworks, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) ListForwarderRoutingPolicies(context.Context, *ttnpb.ListForwarderRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) PublishDownlink(context.Context, *ttnpb.DownlinkMessage) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) PublishUplink(context.Context, *ttnpb.GatewayUplinkMessage) (*pbtypes.Empty, error) {
	return nil, errNotEnabled.New()
}

func (s disabledServer) UpdateGateway(context.Context, *ttnpb.UpdatePacketBrokerGatewayRequest) (*ttnpb.UpdatePacketBrokerGatewayResponse, error) {
	return nil, errNotEnabled.New()
}
