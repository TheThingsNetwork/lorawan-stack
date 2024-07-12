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
	"encoding/binary"
	"net"

	"github.com/google/uuid"
	northboundv1 "go.thethings.industries/pkg/api/gen/tti/gateway/controller/northbound/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/ieee"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func toManagedGateway(
	ids *ttnpb.GatewayIdentifiers, gtw *northboundv1.Gateway, paths []string,
) (*ttnpb.ManagedGateway, error) {
	mapped := &ttnpb.ManagedGateway{
		Ids: ids,
		VersionIds: &ttnpb.GatewayVersionIdentifiers{
			BrandId:         ieee.OUI[gtw.Manufacturer],
			ModelId:         gtw.Model,
			HardwareVersion: gtw.HardwareVersion,
			FirmwareVersion: gtw.FirmwareVersion,
			RuntimeVersion:  gtw.RuntimeVersion,
		},
		CellularImei:       gtw.CellularImei,
		CellularImsi:       gtw.CellularImsi,
		WifiMacAddress:     gtw.WifiMacAddress,
		EthernetMacAddress: gtw.EthernetMacAddress,
	}
	if len(gtw.WifiProfileId) > 0 {
		mapped.WifiProfileId = uuid.Must(uuid.FromBytes(gtw.WifiProfileId)).String()
	}
	if len(gtw.EthernetProfileId) > 0 {
		mapped.EthernetProfileId = uuid.Must(uuid.FromBytes(gtw.EthernetProfileId)).String()
	}
	res := &ttnpb.ManagedGateway{}
	if err := res.SetFields(mapped, paths...); err != nil {
		return nil, err
	}
	return res, nil
}

func toWiFiAccessPoint(ap *northboundv1.WifiAccessPoint) *ttnpb.ManagedGatewayWiFiAccessPoint {
	return &ttnpb.ManagedGatewayWiFiAccessPoint{
		Ssid:               ap.Ssid,
		Bssid:              ap.Bssid,
		Channel:            ap.Channel,
		AuthenticationMode: ap.AuthenticationMode,
		Rssi:               ap.Rssi,
	}
}

var toNetworkInterfaceType = map[northboundv1.NetworkInterfaceType]ttnpb.ManagedGatewayNetworkInterfaceType{
	northboundv1.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_UNSPECIFIED: ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_UNSPECIFIED, //nolint:lll
	northboundv1.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_CELLULAR:    ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_CELLULAR,    //nolint:lll
	northboundv1.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_WIFI:        ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_WIFI,        //nolint:lll
	northboundv1.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_ETHERNET:    ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_ETHERNET,    //nolint:lll
}

var toNetworkInterfaceStatus = map[northboundv1.NetworkInterfaceStatus]ttnpb.ManagedGatewayNetworkInterfaceStatus{
	northboundv1.NetworkInterfaceStatus_NETWORK_INTERFACE_STATUS_UNSPECIFIED: ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UNSPECIFIED, //nolint:lll
	northboundv1.NetworkInterfaceStatus_NETWORK_INTERFACE_STATUS_DOWN:        ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_DOWN,        //nolint:lll
	northboundv1.NetworkInterfaceStatus_NETWORK_INTERFACE_STATUS_UP:          ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UP,          //nolint:lll
	northboundv1.NetworkInterfaceStatus_NETWORK_INTERFACE_STATUS_FAILED:      ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_FAILED,      //nolint:lll
}

func toIPv4(ip uint32) string {
	if ip == 0 {
		return ""
	}
	netIP := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(netIP, ip)
	return netIP.String()
}

func fromIPv4(ip string) uint32 {
	netIP := net.ParseIP(ip).To4()
	if netIP == nil {
		return 0
	}
	return binary.BigEndian.Uint32(netIP)
}

func toNetworkInterfaceAddresses(
	addresses *northboundv1.NetworkInterfaceAddresses,
) *ttnpb.ManagedGatewayNetworkInterfaceAddresses {
	if addresses == nil {
		return nil
	}
	res := &ttnpb.ManagedGatewayNetworkInterfaceAddresses{
		SubnetMask: toIPv4(addresses.SubnetMask),
		Gateway:    toIPv4(addresses.Gateway),
	}
	if len(addresses.IpAddresses) > 0 {
		res.IpAddresses = make([]string, len(addresses.IpAddresses))
		for i, ip := range addresses.IpAddresses {
			res.IpAddresses[i] = toIPv4(ip)
		}
	}
	if len(addresses.DnsServers) > 0 {
		res.DnsServers = make([]string, len(addresses.DnsServers))
		for i, ip := range addresses.DnsServers {
			res.DnsServers[i] = toIPv4(ip)
		}
	}
	return res
}

func fromNetworkInterfaceAddresses(
	addresses *ttnpb.ManagedGatewayNetworkInterfaceAddresses,
) *northboundv1.NetworkInterfaceAddresses {
	if addresses == nil {
		return nil
	}
	res := &northboundv1.NetworkInterfaceAddresses{
		SubnetMask: fromIPv4(addresses.SubnetMask),
		Gateway:    fromIPv4(addresses.Gateway),
	}
	if len(addresses.IpAddresses) > 0 {
		res.IpAddresses = make([]uint32, len(addresses.IpAddresses))
		for i, ip := range addresses.IpAddresses {
			res.IpAddresses[i] = fromIPv4(ip)
		}
	}
	if len(addresses.DnsServers) > 0 {
		res.DnsServers = make([]uint32, len(addresses.DnsServers))
		for i, ip := range addresses.DnsServers {
			res.DnsServers[i] = fromIPv4(ip)
		}
	}
	return res
}

func toNetworkInterfaceInfo(info *northboundv1.NetworkInterfaceInfo) *ttnpb.ManagedGatewayNetworkInterfaceInfo {
	if info == nil {
		return nil
	}
	return &ttnpb.ManagedGatewayNetworkInterfaceInfo{
		Status:      toNetworkInterfaceStatus[info.Status],
		DhcpEnabled: info.DhcpEnabled,
		Addresses:   toNetworkInterfaceAddresses(info.Addresses),
	}
}

func toEvent(
	ids *ttnpb.GatewayIdentifiers, msg *northboundv1.GatewayServiceSubscribeResponse,
) *ttnpb.ManagedGatewayEventData {
	switch update := msg.Update.(type) {
	case *northboundv1.GatewayServiceSubscribeResponse_Gateway:
		gtw, err := toManagedGateway(ids, update.Gateway, ttnpb.ManagedGatewayFieldPathsNested)
		if err != nil {
			return nil
		}
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_Entity{
				Entity: gtw,
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_Location:
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_Location{
				Location: &ttnpb.Location{
					Latitude:  update.Location.Latitude,
					Longitude: update.Location.Longitude,
					Accuracy:  int32(update.Location.Accuracy),
					Source:    ttnpb.LocationSource_SOURCE_WIFI_RSSI_GEOLOCATION,
				},
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_SystemStatus:
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_SystemStatus{
				SystemStatus: &ttnpb.ManagedGatewaySystemStatus{
					CpuTemperature: update.SystemStatus.CpuTemperature,
				},
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_Session:
		networkInterfaceType := ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_UNSPECIFIED
		if update.Session.DisconnectedAt == nil {
			networkInterfaceType = toNetworkInterfaceType[update.Session.NetworkInterfaceType]
		}
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_ControllerConnection{
				ControllerConnection: &ttnpb.ManagedGatewayControllerConnection{
					NetworkInterfaceType: networkInterfaceType,
				},
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_Cellular:
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_CellularBackhaul{
				CellularBackhaul: &ttnpb.ManagedGatewayCellularBackhaul{
					NetworkInterface: toNetworkInterfaceInfo(update.Cellular.NetworkInterface),
					Operator:         update.Cellular.Operator,
					Rssi:             update.Cellular.Rssi,
				},
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_Wifi:
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_WifiBackhaul{
				WifiBackhaul: &ttnpb.ManagedGatewayWiFiBackhaul{
					NetworkInterface:   toNetworkInterfaceInfo(update.Wifi.NetworkInterface),
					Ssid:               update.Wifi.Ssid,
					Bssid:              update.Wifi.Bssid,
					Channel:            update.Wifi.Channel,
					AuthenticationMode: update.Wifi.AuthenticationMode,
					Rssi:               update.Wifi.Rssi,
				},
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_Ethernet:
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_EthernetBackhaul{
				EthernetBackhaul: &ttnpb.ManagedGatewayEthernetBackhaul{
					NetworkInterface: toNetworkInterfaceInfo(update.Ethernet.NetworkInterface),
				},
			},
		}
	case *northboundv1.GatewayServiceSubscribeResponse_LoraPacketForwarder:
		return &ttnpb.ManagedGatewayEventData{
			Data: &ttnpb.ManagedGatewayEventData_GatewayServerConnection{
				GatewayServerConnection: &ttnpb.ManagedGatewayGatewayServerConnection{
					NetworkInterfaceType: toNetworkInterfaceType[update.LoraPacketForwarder.NetworkInterfaceType],
					Address:              update.LoraPacketForwarder.Address,
				},
			},
		}
	}
	return nil
}

func toProfileID(profileID []byte) string {
	return uuid.Must(uuid.FromBytes(profileID)).String()
}

func fromProfileID(profileID string) []byte {
	id := uuid.Must(uuid.Parse(profileID))
	return id[:]
}

func fromProfileIDOrNil(profileID string) *northboundv1.ProfileIDValue {
	if profileID == "" {
		return nil
	}
	id := uuid.Must(uuid.Parse(profileID))
	return &northboundv1.ProfileIDValue{
		Value: id[:],
	}
}

func toWiFiProfile(profileID []byte, profile *northboundv1.WifiProfile) *ttnpb.ManagedGatewayWiFiProfile {
	if profile == nil {
		return nil
	}
	return &ttnpb.ManagedGatewayWiFiProfile{
		ProfileId:                 toProfileID(profileID),
		ProfileName:               profile.ProfileName,
		Ssid:                      profile.Ssid,
		Password:                  profile.Password,
		NetworkInterfaceAddresses: toNetworkInterfaceAddresses(profile.NetworkInterfaceAddresses),
	}
}

func fromWiFiProfile(profile *ttnpb.ManagedGatewayWiFiProfile) *northboundv1.WifiProfile {
	if profile == nil {
		return nil
	}
	return &northboundv1.WifiProfile{
		ProfileName:               profile.ProfileName,
		Shared:                    true,
		Ssid:                      profile.Ssid,
		Password:                  profile.Password,
		NetworkInterfaceAddresses: fromNetworkInterfaceAddresses(profile.NetworkInterfaceAddresses),
	}
}

func toEthernetProfile(profileID []byte, profile *northboundv1.EthernetProfile) *ttnpb.ManagedGatewayEthernetProfile {
	if profile == nil {
		return nil
	}
	return &ttnpb.ManagedGatewayEthernetProfile{
		ProfileId:                 toProfileID(profileID),
		ProfileName:               profile.ProfileName,
		NetworkInterfaceAddresses: toNetworkInterfaceAddresses(profile.NetworkInterfaceAddresses),
	}
}

func fromEthernetProfile(profile *ttnpb.ManagedGatewayEthernetProfile) *northboundv1.EthernetProfile {
	if profile == nil {
		return nil
	}
	return &northboundv1.EthernetProfile{
		ProfileName:               profile.ProfileName,
		Shared:                    true,
		NetworkInterfaceAddresses: fromNetworkInterfaceAddresses(profile.NetworkInterfaceAddresses),
	}
}
