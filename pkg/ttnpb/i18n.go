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

	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
)

func defineEnum(e fmt.Stringer, message string) {
	i18n.Define("enum:"+e.String(), message).SetSource(1)
}

func init() {
	defineEnum(GrantType_GRANT_AUTHORIZATION_CODE, "authorization code")
	defineEnum(GrantType_GRANT_PASSWORD, "username and password")
	defineEnum(GrantType_GRANT_REFRESH_TOKEN, "refresh token")

	defineEnum(State_STATE_REQUESTED, "requested and pending review")
	defineEnum(State_STATE_APPROVED, "reviewed and approved")
	defineEnum(State_STATE_REJECTED, "reviewed and rejected")
	defineEnum(State_STATE_FLAGGED, "flagged and pending review")
	defineEnum(State_STATE_SUSPENDED, "reviewed and suspended")

	defineEnum(ContactType_CONTACT_TYPE_OTHER, "other")
	defineEnum(ContactType_CONTACT_TYPE_ABUSE, "abuse")
	defineEnum(ContactType_CONTACT_TYPE_BILLING, "billing")
	defineEnum(ContactType_CONTACT_TYPE_TECHNICAL, "technical")

	defineEnum(ContactMethod_CONTACT_METHOD_OTHER, "other")
	defineEnum(ContactMethod_CONTACT_METHOD_EMAIL, "email")
	defineEnum(ContactMethod_CONTACT_METHOD_PHONE, "phone")

	defineEnum(MType_JOIN_REQUEST, "join request")
	defineEnum(MType_JOIN_ACCEPT, "join accept")
	defineEnum(MType_UNCONFIRMED_UP, "unconfirmed uplink")
	defineEnum(MType_UNCONFIRMED_DOWN, "unconfirmed downlink")
	defineEnum(MType_CONFIRMED_UP, "confirmed uplink")
	defineEnum(MType_CONFIRMED_DOWN, "confirmed downlink")
	defineEnum(MType_REJOIN_REQUEST, "rejoin request")
	defineEnum(MType_PROPRIETARY, "proprietary")

	defineEnum(JoinRequestType_REJOIN_CONTEXT, "rejoin to renew context")
	defineEnum(JoinRequestType_REJOIN_SESSION, "rejoin to renew session")
	defineEnum(JoinRequestType_REJOIN_KEYS, "rejoin to renew keys")
	defineEnum(JoinRequestType_JOIN, "join")

	defineEnum(RejoinRequestType_CONTEXT, "renew context")
	defineEnum(RejoinRequestType_SESSION, "renew session")
	defineEnum(RejoinRequestType_KEYS, "renew keys")

	defineEnum(CFListType_FREQUENCIES, "frequencies")
	defineEnum(CFListType_CHANNEL_MASKS, "channel masks")

	defineEnum(MACCommandIdentifier_CID_RFU_0, "RFU")
	defineEnum(MACCommandIdentifier_CID_RESET, "reset")
	defineEnum(MACCommandIdentifier_CID_LINK_CHECK, "link check")
	defineEnum(MACCommandIdentifier_CID_LINK_ADR, "adaptive data rate")
	defineEnum(MACCommandIdentifier_CID_DUTY_CYCLE, "duty cycle")
	defineEnum(MACCommandIdentifier_CID_RX_PARAM_SETUP, "receive parameters")
	defineEnum(MACCommandIdentifier_CID_DEV_STATUS, "device status")
	defineEnum(MACCommandIdentifier_CID_NEW_CHANNEL, "new channel")
	defineEnum(MACCommandIdentifier_CID_RX_TIMING_SETUP, "receive timing")
	defineEnum(MACCommandIdentifier_CID_TX_PARAM_SETUP, "transmit parameters")
	defineEnum(MACCommandIdentifier_CID_DL_CHANNEL, "downlink channel")
	defineEnum(MACCommandIdentifier_CID_REKEY, "rekey")
	defineEnum(MACCommandIdentifier_CID_ADR_PARAM_SETUP, "adaptive data rate parameters")
	defineEnum(MACCommandIdentifier_CID_DEVICE_TIME, "device time")
	defineEnum(MACCommandIdentifier_CID_FORCE_REJOIN, "force rejoin")
	defineEnum(MACCommandIdentifier_CID_REJOIN_PARAM_SETUP, "rejoin parameters")
	defineEnum(MACCommandIdentifier_CID_PING_SLOT_INFO, "ping slot info")
	defineEnum(MACCommandIdentifier_CID_PING_SLOT_CHANNEL, "ping slot channel")
	defineEnum(MACCommandIdentifier_CID_BEACON_TIMING, "beacon timing")
	defineEnum(MACCommandIdentifier_CID_BEACON_FREQ, "beacon frequency")
	defineEnum(MACCommandIdentifier_CID_DEVICE_MODE, "device mode")

	defineEnum(MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF, "end device relay configuration")
	defineEnum(MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST, "update uplink forwarding rules")
	defineEnum(MACCommandIdentifier_CID_RELAY_CONF, "relay configuration")
	defineEnum(MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST, "manage uplink forwarding rules")
	defineEnum(MACCommandIdentifier_CID_RELAY_NOTIFY_NEW_END_DEVICE, "new end device under relay")
	defineEnum(MACCommandIdentifier_CID_RELAY_FILTER_LIST, "manage join request forwarding rules")
	defineEnum(MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT, "manage uplink forwarding limits")

	defineEnum(LocationSource_SOURCE_UNKNOWN, "unknown location source")
	defineEnum(LocationSource_SOURCE_GPS, "determined by GPS")
	defineEnum(LocationSource_SOURCE_REGISTRY, "set in and updated from a registry")
	defineEnum(LocationSource_SOURCE_IP_GEOLOCATION, "estimated with IP geolocation")
	defineEnum(LocationSource_SOURCE_WIFI_RSSI_GEOLOCATION, "estimated with WiFi RSSI geolocation")
	defineEnum(LocationSource_SOURCE_BT_RSSI_GEOLOCATION, "estimated with Bluetooth RSSI geolocation")
	defineEnum(LocationSource_SOURCE_LORA_RSSI_GEOLOCATION, "estimated with LoRa RSSI geolocation")
	defineEnum(LocationSource_SOURCE_LORA_TDOA_GEOLOCATION, "estimated with LoRa TDOA geolocation")
	defineEnum(LocationSource_SOURCE_COMBINED_GEOLOCATION, "estimated by a combination of geolocation sources")

	defineEnum(PayloadFormatter_FORMATTER_NONE, "no formatter")
	defineEnum(PayloadFormatter_FORMATTER_REPOSITORY, "defined by end device type repository")
	defineEnum(PayloadFormatter_FORMATTER_GRPC_SERVICE, "gRPC service")
	defineEnum(PayloadFormatter_FORMATTER_JAVASCRIPT, "JavaScript")
	defineEnum(PayloadFormatter_FORMATTER_CAYENNELPP, "Cayenne LPP")

	defineEnum(Right_RIGHT_USER_INFO, "view user information")
	defineEnum(Right_RIGHT_USER_SETTINGS_BASIC, "edit basic user settings")
	defineEnum(Right_RIGHT_USER_SETTINGS_API_KEYS, "view and edit user API keys")
	defineEnum(Right_RIGHT_USER_DELETE, "delete user account")
	defineEnum(Right_RIGHT_USER_AUTHORIZED_CLIENTS, "view and edit authorized OAuth clients of the user")
	defineEnum(Right_RIGHT_USER_APPLICATIONS_LIST, "list applications the user is a collaborator of")
	defineEnum(Right_RIGHT_USER_APPLICATIONS_CREATE, "create an application under the user account")
	defineEnum(Right_RIGHT_USER_GATEWAYS_LIST, "list gateways the user is a collaborator of")
	defineEnum(Right_RIGHT_USER_GATEWAYS_CREATE, "create a gateway under the user account")
	defineEnum(Right_RIGHT_USER_CLIENTS_LIST, "list OAuth clients the user is a collaborator of")
	defineEnum(Right_RIGHT_USER_CLIENTS_CREATE, "create an OAuth client under the user account")
	defineEnum(Right_RIGHT_USER_ORGANIZATIONS_LIST, "list organizations the user is a member of")
	defineEnum(Right_RIGHT_USER_ORGANIZATIONS_CREATE, "create an organization under the user account")
	defineEnum(Right_RIGHT_USER_NOTIFICATIONS_READ, "read user notifications")
	defineEnum(Right_RIGHT_USER_ALL, "all user rights")

	defineEnum(Right_RIGHT_APPLICATION_INFO, "view application information")
	defineEnum(Right_RIGHT_APPLICATION_SETTINGS_BASIC, "edit basic application settings")
	defineEnum(Right_RIGHT_APPLICATION_SETTINGS_API_KEYS, "view and edit application API keys")
	defineEnum(Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS, "view and edit application collaborators")
	defineEnum(Right_RIGHT_APPLICATION_SETTINGS_PACKAGES, "view and edit application packages and associations")
	defineEnum(Right_RIGHT_APPLICATION_DELETE, "delete application")
	defineEnum(Right_RIGHT_APPLICATION_DEVICES_READ, "view devices in application")
	defineEnum(Right_RIGHT_APPLICATION_DEVICES_WRITE, "create devices in application")
	defineEnum(Right_RIGHT_APPLICATION_DEVICES_READ_KEYS, "view device keys in application")
	defineEnum(Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS, "edit device keys in application")
	defineEnum(Right_RIGHT_APPLICATION_TRAFFIC_READ, "read application traffic (uplink and downlink)")
	defineEnum(Right_RIGHT_APPLICATION_TRAFFIC_UP_WRITE, "write uplink application traffic")
	defineEnum(Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE, "write downlink application traffic")
	defineEnum(Right_RIGHT_APPLICATION_LINK, "link as Application to a Network Server for traffic exchange, i.e. read uplink and write downlink")
	defineEnum(Right_RIGHT_APPLICATION_ALL, "all application rights")

	defineEnum(Right_RIGHT_CLIENT_ALL, "all OAuth client rights")
	defineEnum(Right_RIGHT_CLIENT_INFO, "view OAuth client information")
	defineEnum(Right_RIGHT_CLIENT_SETTINGS_BASIC, "edit OAuth client basic settings")
	defineEnum(Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS, "view and edit OAuth client collaborators")
	defineEnum(Right_RIGHT_CLIENT_DELETE, "delete OAuth client")

	defineEnum(Right_RIGHT_GATEWAY_INFO, "view gateway information")
	defineEnum(Right_RIGHT_GATEWAY_SETTINGS_BASIC, "edit basic gateway settings")
	defineEnum(Right_RIGHT_GATEWAY_SETTINGS_API_KEYS, "view and edit gateway API keys")
	defineEnum(Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS, "view and edit gateway collaborators")
	defineEnum(Right_RIGHT_GATEWAY_DELETE, "delete gateway")
	defineEnum(Right_RIGHT_GATEWAY_TRAFFIC_READ, "read gateway traffic")
	defineEnum(Right_RIGHT_GATEWAY_TRAFFIC_DOWN_WRITE, "write downlink gateway traffic")
	defineEnum(Right_RIGHT_GATEWAY_LINK, "link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink")
	defineEnum(Right_RIGHT_GATEWAY_STATUS_READ, "view gateway status")
	defineEnum(Right_RIGHT_GATEWAY_LOCATION_READ, "view gateway location")
	defineEnum(Right_RIGHT_GATEWAY_WRITE_SECRETS, "store secrets for a gateway")
	defineEnum(Right_RIGHT_GATEWAY_READ_SECRETS, "retrieve secrets associated with a gateway")
	defineEnum(Right_RIGHT_GATEWAY_ALL, "all gateway rights")

	defineEnum(Right_RIGHT_ORGANIZATION_INFO, "view organization information")
	defineEnum(Right_RIGHT_ORGANIZATION_SETTINGS_BASIC, "edit basic organization settings")
	defineEnum(Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS, "view and edit organization API keys")
	defineEnum(Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS, "view and edit organization members")
	defineEnum(Right_RIGHT_ORGANIZATION_DELETE, "delete organization")
	defineEnum(Right_RIGHT_ORGANIZATION_APPLICATIONS_LIST, "list the applications the organization is a collaborator of")
	defineEnum(Right_RIGHT_ORGANIZATION_APPLICATIONS_CREATE, "create an application under the organization")
	defineEnum(Right_RIGHT_ORGANIZATION_GATEWAYS_LIST, "list the gateways the organization is a collaborator of")
	defineEnum(Right_RIGHT_ORGANIZATION_GATEWAYS_CREATE, "create a gateway under the organization")
	defineEnum(Right_RIGHT_ORGANIZATION_CLIENTS_LIST, "list the OAuth clients the organization is a collaborator of")
	defineEnum(Right_RIGHT_ORGANIZATION_CLIENTS_CREATE, "create an OAuth client under the organization")
	defineEnum(Right_RIGHT_ORGANIZATION_ADD_AS_COLLABORATOR, "add the organization as a collaborator on an existing entity")
	defineEnum(Right_RIGHT_ORGANIZATION_ALL, "all organization rights")

	defineEnum(Right_RIGHT_SEND_INVITES, "send user invites")

	defineEnum(Right_RIGHT_ALL, "all possible rights")
}
