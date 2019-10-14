// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import (
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/i18n"
)

func defineEnum(e fmt.Stringer, message string) {
	i18n.Define("enum:"+e.String(), message).SetSource(1)
}

func init() {
	defineEnum(GRANT_AUTHORIZATION_CODE, "authorization code")
	defineEnum(GRANT_PASSWORD, "username and password")
	defineEnum(GRANT_REFRESH_TOKEN, "refresh token")

	defineEnum(STATE_REQUESTED, "requested and pending review")
	defineEnum(STATE_APPROVED, "reviewed and approved")
	defineEnum(STATE_REJECTED, "reviewed and rejected")
	defineEnum(STATE_FLAGGED, "flagged and pending review")
	defineEnum(STATE_SUSPENDED, "reviewed and suspended")

	defineEnum(CONTACT_TYPE_OTHER, "other")
	defineEnum(CONTACT_TYPE_ABUSE, "abuse")
	defineEnum(CONTACT_TYPE_BILLING, "billing")
	defineEnum(CONTACT_TYPE_TECHNICAL, "technical")

	defineEnum(CONTACT_METHOD_OTHER, "other")
	defineEnum(CONTACT_METHOD_EMAIL, "email")
	defineEnum(CONTACT_METHOD_PHONE, "phone")

	defineEnum(MType_JOIN_REQUEST, "join request")
	defineEnum(MType_JOIN_ACCEPT, "join accept")
	defineEnum(MType_UNCONFIRMED_UP, "unconfirmed uplink")
	defineEnum(MType_UNCONFIRMED_DOWN, "unconfirmed downlink")
	defineEnum(MType_CONFIRMED_UP, "confirmed uplink")
	defineEnum(MType_CONFIRMED_DOWN, "confirmed downlink")
	defineEnum(MType_REJOIN_REQUEST, "rejoin request")
	defineEnum(MType_PROPRIETARY, "proprietary")

	defineEnum(RejoinType_CONTEXT, "renew context")
	defineEnum(RejoinType_SESSION, "renew session")
	defineEnum(RejoinType_KEYS, "renew keys")

	defineEnum(CFListType_FREQUENCIES, "frequencies")
	defineEnum(CFListType_CHANNEL_MASKS, "channel masks")

	defineEnum(CID_RFU_0, "RFU")
	defineEnum(CID_RESET, "reset")
	defineEnum(CID_LINK_CHECK, "link check")
	defineEnum(CID_LINK_ADR, "adaptive data rate")
	defineEnum(CID_DUTY_CYCLE, "duty cycle")
	defineEnum(CID_RX_PARAM_SETUP, "receive parameters")
	defineEnum(CID_DEV_STATUS, "device status")
	defineEnum(CID_NEW_CHANNEL, "new channel")
	defineEnum(CID_RX_TIMING_SETUP, "receive timing")
	defineEnum(CID_TX_PARAM_SETUP, "transmit parameters")
	defineEnum(CID_DL_CHANNEL, "downlink channel")
	defineEnum(CID_REKEY, "rekey")
	defineEnum(CID_ADR_PARAM_SETUP, "adaptive data rate parameters")
	defineEnum(CID_DEVICE_TIME, "device time")
	defineEnum(CID_FORCE_REJOIN, "force rejoin")
	defineEnum(CID_REJOIN_PARAM_SETUP, "rejoin parameters")
	defineEnum(CID_PING_SLOT_INFO, "ping slot info")
	defineEnum(CID_PING_SLOT_CHANNEL, "ping slot channel")
	defineEnum(CID_BEACON_TIMING, "beacon timing")
	defineEnum(CID_BEACON_FREQ, "beacon frequency")
	defineEnum(CID_DEVICE_MODE, "device mode")

	defineEnum(SOURCE_UNKNOWN, "unknown location source")
	defineEnum(SOURCE_GPS, "determined by GPS")
	defineEnum(SOURCE_REGISTRY, "set in and updated from a registry")
	defineEnum(SOURCE_IP_GEOLOCATION, "estimated with IP geolocation")
	defineEnum(SOURCE_WIFI_RSSI_GEOLOCATION, "estimated with WiFi RSSI geolocation")
	defineEnum(SOURCE_BT_RSSI_GEOLOCATION, "estimated with Bluetooth RSSI geolocation")
	defineEnum(SOURCE_LORA_RSSI_GEOLOCATION, "estimated with LoRa RSSI geolocation")
	defineEnum(SOURCE_LORA_TDOA_GEOLOCATION, "estimated with LoRa TDOA geolocation")
	defineEnum(SOURCE_COMBINED_GEOLOCATION, "estimated by a combination of geolocation sources")

	defineEnum(PayloadFormatter_FORMATTER_NONE, "no formatter")
	defineEnum(PayloadFormatter_FORMATTER_REPOSITORY, "defined by end device type repository")
	defineEnum(PayloadFormatter_FORMATTER_GRPC_SERVICE, "gRPC service")
	defineEnum(PayloadFormatter_FORMATTER_JAVASCRIPT, "JavaScript")
	defineEnum(PayloadFormatter_FORMATTER_CAYENNELPP, "Cayenne LPP")

	defineEnum(RIGHT_USER_INFO, "view user information")
	defineEnum(RIGHT_USER_SETTINGS_BASIC, "edit basic user settings")
	defineEnum(RIGHT_USER_SETTINGS_API_KEYS, "view and edit user API keys")
	defineEnum(RIGHT_USER_DELETE, "delete user account")
	defineEnum(RIGHT_USER_AUTHORIZED_CLIENTS, "view and edit authorized OAuth clients of the user")
	defineEnum(RIGHT_USER_APPLICATIONS_LIST, "list applications the user is a collaborator of")
	defineEnum(RIGHT_USER_APPLICATIONS_CREATE, "create an application under the user account")
	defineEnum(RIGHT_USER_GATEWAYS_LIST, "list gateways the user is a collaborator of")
	defineEnum(RIGHT_USER_GATEWAYS_CREATE, "create a gateway under the user account")
	defineEnum(RIGHT_USER_CLIENTS_LIST, "list OAuth clients the user is a collaborator of")
	defineEnum(RIGHT_USER_CLIENTS_CREATE, "create an OAuth client under the user account")
	defineEnum(RIGHT_USER_ORGANIZATIONS_LIST, "list organizations the user is a member of")
	defineEnum(RIGHT_USER_ORGANIZATIONS_CREATE, "create an organization under the user account")
	defineEnum(RIGHT_USER_ALL, "all user rights")

	defineEnum(RIGHT_APPLICATION_INFO, "view application information")
	defineEnum(RIGHT_APPLICATION_SETTINGS_BASIC, "edit basic application settings")
	defineEnum(RIGHT_APPLICATION_SETTINGS_API_KEYS, "view and edit application API keys")
	defineEnum(RIGHT_APPLICATION_SETTINGS_COLLABORATORS, "view and edit application collaborators")
	defineEnum(RIGHT_APPLICATION_SETTINGS_PACKAGES, "view and edit application packages and associations")
	defineEnum(RIGHT_APPLICATION_DELETE, "delete application")
	defineEnum(RIGHT_APPLICATION_DEVICES_READ, "view devices in application")
	defineEnum(RIGHT_APPLICATION_DEVICES_WRITE, "create devices in application")
	defineEnum(RIGHT_APPLICATION_DEVICES_READ_KEYS, "view device keys in application")
	defineEnum(RIGHT_APPLICATION_DEVICES_WRITE_KEYS, "edit device keys in application")
	defineEnum(RIGHT_APPLICATION_TRAFFIC_READ, "read application traffic (uplink and downlink)")
	defineEnum(RIGHT_APPLICATION_TRAFFIC_UP_WRITE, "write uplink application traffic")
	defineEnum(RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE, "write downlink application traffic")
	defineEnum(RIGHT_APPLICATION_LINK, "link as Application to a Network Server for traffic exchange, i.e. read uplink and write downlink")
	defineEnum(RIGHT_APPLICATION_ALL, "all application rights")

	defineEnum(RIGHT_CLIENT_ALL, "all OAuth client rights")

	defineEnum(RIGHT_GATEWAY_INFO, "view gateway information")
	defineEnum(RIGHT_GATEWAY_SETTINGS_BASIC, "edit basic gateway settings")
	defineEnum(RIGHT_GATEWAY_SETTINGS_API_KEYS, "view and edit gateway API keys")
	defineEnum(RIGHT_GATEWAY_SETTINGS_COLLABORATORS, "view and edit gateway collaborators")
	defineEnum(RIGHT_GATEWAY_DELETE, "delete gateway")
	defineEnum(RIGHT_GATEWAY_TRAFFIC_READ, "read gateway traffic")
	defineEnum(RIGHT_GATEWAY_TRAFFIC_DOWN_WRITE, "write downlink gateway traffic")
	defineEnum(RIGHT_GATEWAY_LINK, "link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink")
	defineEnum(RIGHT_GATEWAY_STATUS_READ, "view gateway status")
	defineEnum(RIGHT_GATEWAY_LOCATION_READ, "view gateway location")
	defineEnum(RIGHT_GATEWAY_ALL, "all gateway rights")

	defineEnum(RIGHT_ORGANIZATION_INFO, "view organization information")
	defineEnum(RIGHT_ORGANIZATION_SETTINGS_BASIC, "edit basic organization settings")
	defineEnum(RIGHT_ORGANIZATION_SETTINGS_API_KEYS, "view and edit organization API keys")
	defineEnum(RIGHT_ORGANIZATION_SETTINGS_MEMBERS, "view and edit organization members")
	defineEnum(RIGHT_ORGANIZATION_DELETE, "delete organization")
	defineEnum(RIGHT_ORGANIZATION_APPLICATIONS_LIST, "list the applications the organization is a collaborator of")
	defineEnum(RIGHT_ORGANIZATION_APPLICATIONS_CREATE, "create an application under the organization")
	defineEnum(RIGHT_ORGANIZATION_GATEWAYS_LIST, "list the gateways the organization is a collaborator of")
	defineEnum(RIGHT_ORGANIZATION_GATEWAYS_CREATE, "create a gateway under the organization")
	defineEnum(RIGHT_ORGANIZATION_CLIENTS_LIST, "list the OAuth clients the organization is a collaborator of")
	defineEnum(RIGHT_ORGANIZATION_CLIENTS_CREATE, "create an OAuth client under the organization")
	defineEnum(RIGHT_ORGANIZATION_ADD_AS_COLLABORATOR, "add the organization as a collaborator on an existing entity")
	defineEnum(RIGHT_ORGANIZATION_ALL, "all organization rights")

	defineEnum(RIGHT_SEND_INVITES, "send user invites")

	defineEnum(RIGHT_ALL, "all possible rights")
}
