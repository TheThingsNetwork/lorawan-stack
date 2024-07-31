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

package client

import (
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EventNamePattern is the pattern for event names published by the managed gateway client.
const EventNamePattern = `/^gcs\.managed\..*/`

var (
	evtUpdateManagedGateway = events.Define(
		"gcs.managed.update", "update managed gateway",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGateway{}),
	)
	evtUpdateManagedGatewayLocation = events.Define(
		"gcs.managed.location.update", "update location",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.Location{}),
	)
	evtReceiveManagedGatewaySystemStatus = events.Define(
		"gcs.managed.system_status.receive", "receive system status",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGatewaySystemStatus{}),
	)
	evtManagedGatewayControllerUp = events.Define(
		"gcs.managed.controller.up", "controller connection up",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGatewayControllerConnection{}),
	)
	evtManagedGatewayControllerDown = events.Define(
		"gcs.managed.controller.down", "controller connection down",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
	)
	evtManagedGatewayGatewayServerUp = events.Define(
		"gcs.managed.gs.up", "connection with Gateway Server up",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGatewayGatewayServerConnection{}),
	)
	evtManagedGatewayGatewayServerDown = events.Define(
		"gcs.managed.gs.down", "connection with Gateway Server down",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
	)
	evtManagedGatewayCellularUp = events.Define(
		"gcs.managed.cellular.up", "cellular backhaul up",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGatewayCellularBackhaul{}),
	)
	evtManagedGatewayCellularDown = events.Define(
		"gcs.managed.cellular.down", "cellular backhaul down",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
	)
	evtManagedGatewayWiFiUp = events.Define(
		"gcs.managed.wifi.up", "WiFi backhaul up",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGatewayWiFiBackhaul{}),
	)
	evtManagedGatewayWiFiDown = events.Define(
		"gcs.managed.wifi.down", "WiFi backhaul down",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
	)
	evtManagedGatewayWiFiFail = events.Define(
		"gcs.managed.wifi.fail", "WiFi backhaul fail",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
	)
	evtManagedGatewayEthernetUp = events.Define(
		"gcs.managed.ethernet.up", "Ethernet backhaul up",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithDataType(&ttnpb.ManagedGatewayEthernetBackhaul{}),
	)
	evtManagedGatewayEthernetDown = events.Define(
		"gcs.managed.ethernet.down", "Ethernet backhaul down",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
	)
)
