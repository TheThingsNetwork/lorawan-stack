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
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func asDevAddrBlocks(blocks []*packetbroker.DevAddrBlock) []*ttnpb.PacketBrokerDevAddrBlock {
	res := make([]*ttnpb.PacketBrokerDevAddrBlock, len(blocks))
	for i, b := range blocks {
		res[i] = &ttnpb.PacketBrokerDevAddrBlock{
			DevAddrPrefix: &ttnpb.DevAddrPrefix{
				DevAddr: &types.DevAddr{},
				Length:  b.GetPrefix().GetLength(),
			},
			HomeNetworkClusterID: b.GetHomeNetworkClusterId(),
		}
		res[i].DevAddrPrefix.DevAddr.UnmarshalNumber(b.GetPrefix().GetValue())
	}
	return res
}

func toDevAddrBlocks(blocks []*ttnpb.PacketBrokerDevAddrBlock) []*packetbroker.DevAddrBlock {
	res := make([]*packetbroker.DevAddrBlock, len(blocks))
	for i, b := range blocks {
		res[i] = &packetbroker.DevAddrBlock{
			Prefix: &packetbroker.DevAddrPrefix{
				Value:  b.GetDevAddrPrefix().DevAddr.MarshalNumber(),
				Length: b.GetDevAddrPrefix().GetLength(),
			},
			HomeNetworkClusterId: b.GetHomeNetworkClusterID(),
		}
	}
	return res
}

func asContactInfo(admin, technical *packetbroker.ContactInfo) []*ttnpb.ContactInfo {
	res := make([]*ttnpb.ContactInfo, 0, 2)
	if email := admin.GetEmail(); email != "" {
		res = append(res, &ttnpb.ContactInfo{
			ContactType:   ttnpb.CONTACT_TYPE_OTHER,
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         email,
		})
	}
	if email := technical.GetEmail(); email != "" {
		res = append(res, &ttnpb.ContactInfo{
			ContactType:   ttnpb.CONTACT_TYPE_TECHNICAL,
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         email,
		})
	}
	return res
}

func toContactInfo(info []*ttnpb.ContactInfo) (admin, technical *packetbroker.ContactInfo) {
	for _, c := range info {
		if c.GetContactMethod() != ttnpb.CONTACT_METHOD_EMAIL || c.GetValue() == "" {
			continue
		}
		switch c.GetContactType() {
		case ttnpb.CONTACT_TYPE_OTHER:
			admin = &packetbroker.ContactInfo{
				Email: c.GetValue(),
			}
		case ttnpb.CONTACT_TYPE_TECHNICAL:
			technical = &packetbroker.ContactInfo{
				Email: c.GetValue(),
			}
		}
	}
	return
}

func asUplinkRoutingPolicy(policy *packetbroker.RoutingPolicy_Uplink) *ttnpb.PacketBrokerRoutingPolicyUplink {
	return &ttnpb.PacketBrokerRoutingPolicyUplink{
		JoinRequest:     policy.GetJoinRequest(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
		SignalQuality:   policy.GetSignalQuality(),
		Localization:    policy.GetLocalization(),
	}
}

func asDownlinkRoutingPolicy(policy *packetbroker.RoutingPolicy_Downlink) *ttnpb.PacketBrokerRoutingPolicyDownlink {
	return &ttnpb.PacketBrokerRoutingPolicyDownlink{
		JoinAccept:      policy.GetJoinAccept(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
	}
}

func asDefaultRoutingPolicy(policy *packetbroker.RoutingPolicy) *ttnpb.PacketBrokerDefaultRoutingPolicy {
	return &ttnpb.PacketBrokerDefaultRoutingPolicy{
		UpdatedAt: policy.GetUpdatedAt(),
		Uplink:    asUplinkRoutingPolicy(policy.GetUplink()),
		Downlink:  asDownlinkRoutingPolicy(policy.GetDownlink()),
	}
}

func asRoutingPolicy(policy *packetbroker.RoutingPolicy) *ttnpb.PacketBrokerRoutingPolicy {
	var homeNetworkID *ttnpb.PacketBrokerNetworkIdentifier
	if policy.HomeNetworkNetId != 0 || policy.HomeNetworkTenantId != "" {
		homeNetworkID = &ttnpb.PacketBrokerNetworkIdentifier{
			NetID:    policy.GetHomeNetworkNetId(),
			TenantID: policy.GetHomeNetworkTenantId(),
		}
	}
	return &ttnpb.PacketBrokerRoutingPolicy{
		ForwarderId: &ttnpb.PacketBrokerNetworkIdentifier{
			NetID:    policy.GetForwarderNetId(),
			TenantID: policy.GetForwarderTenantId(),
		},
		HomeNetworkId: homeNetworkID,
		UpdatedAt:     policy.GetUpdatedAt(),
		Uplink:        asUplinkRoutingPolicy(policy.GetUplink()),
		Downlink:      asDownlinkRoutingPolicy(policy.GetDownlink()),
	}
}

func toUplinkRoutingPolicy(policy *ttnpb.PacketBrokerRoutingPolicyUplink) *packetbroker.RoutingPolicy_Uplink {
	return &packetbroker.RoutingPolicy_Uplink{
		JoinRequest:     policy.GetJoinRequest(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
		SignalQuality:   policy.GetSignalQuality(),
		Localization:    policy.GetLocalization(),
	}
}

func toDownlinkRoutingPolicy(policy *ttnpb.PacketBrokerRoutingPolicyDownlink) *packetbroker.RoutingPolicy_Downlink {
	return &packetbroker.RoutingPolicy_Downlink{
		JoinAccept:      policy.GetJoinAccept(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
	}
}
