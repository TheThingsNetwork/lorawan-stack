// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttnpb

var PacketBrokerNetworkIdentifierFieldPathsNested = []string{
	"net_id",
	"tenant_id",
}

var PacketBrokerNetworkIdentifierFieldPathsTopLevel = []string{
	"net_id",
	"tenant_id",
}
var PacketBrokerDevAddrBlockFieldPathsNested = []string{
	"dev_addr_prefix",
	"dev_addr_prefix.dev_addr",
	"dev_addr_prefix.length",
	"home_network_cluster_id",
}

var PacketBrokerDevAddrBlockFieldPathsTopLevel = []string{
	"dev_addr_prefix",
	"home_network_cluster_id",
}
var PacketBrokerNetworkFieldPathsNested = []string{
	"contact_info",
	"dev_addr_blocks",
	"id",
	"id.net_id",
	"id.tenant_id",
	"name",
}

var PacketBrokerNetworkFieldPathsTopLevel = []string{
	"contact_info",
	"dev_addr_blocks",
	"id",
	"name",
}
var PacketBrokerNetworksFieldPathsNested = []string{
	"networks",
}

var PacketBrokerNetworksFieldPathsTopLevel = []string{
	"networks",
}
var PacketBrokerInfoFieldPathsNested = []string{
	"forwarder_enabled",
	"home_network_enabled",
	"registration",
	"registration.contact_info",
	"registration.dev_addr_blocks",
	"registration.id",
	"registration.id.net_id",
	"registration.id.tenant_id",
	"registration.name",
}

var PacketBrokerInfoFieldPathsTopLevel = []string{
	"forwarder_enabled",
	"home_network_enabled",
	"registration",
}
var PacketBrokerRoutingPolicyUplinkFieldPathsNested = []string{
	"application_data",
	"join_request",
	"localization",
	"mac_data",
	"signal_quality",
}

var PacketBrokerRoutingPolicyUplinkFieldPathsTopLevel = []string{
	"application_data",
	"join_request",
	"localization",
	"mac_data",
	"signal_quality",
}
var PacketBrokerRoutingPolicyDownlinkFieldPathsNested = []string{
	"application_data",
	"join_accept",
	"mac_data",
}

var PacketBrokerRoutingPolicyDownlinkFieldPathsTopLevel = []string{
	"application_data",
	"join_accept",
	"mac_data",
}
var PacketBrokerDefaultRoutingPolicyFieldPathsNested = []string{
	"downlink",
	"downlink.application_data",
	"downlink.join_accept",
	"downlink.mac_data",
	"updated_at",
	"uplink",
	"uplink.application_data",
	"uplink.join_request",
	"uplink.localization",
	"uplink.mac_data",
	"uplink.signal_quality",
}

var PacketBrokerDefaultRoutingPolicyFieldPathsTopLevel = []string{
	"downlink",
	"updated_at",
	"uplink",
}
var PacketBrokerRoutingPolicyFieldPathsNested = []string{
	"downlink",
	"downlink.application_data",
	"downlink.join_accept",
	"downlink.mac_data",
	"forwarder_id",
	"forwarder_id.net_id",
	"forwarder_id.tenant_id",
	"home_network_id",
	"home_network_id.net_id",
	"home_network_id.tenant_id",
	"updated_at",
	"uplink",
	"uplink.application_data",
	"uplink.join_request",
	"uplink.localization",
	"uplink.mac_data",
	"uplink.signal_quality",
}

var PacketBrokerRoutingPolicyFieldPathsTopLevel = []string{
	"downlink",
	"forwarder_id",
	"home_network_id",
	"updated_at",
	"uplink",
}
var SetPacketBrokerDefaultRoutingPolicyRequestFieldPathsNested = []string{
	"downlink",
	"downlink.application_data",
	"downlink.join_accept",
	"downlink.mac_data",
	"uplink",
	"uplink.application_data",
	"uplink.join_request",
	"uplink.localization",
	"uplink.mac_data",
	"uplink.signal_quality",
}

var SetPacketBrokerDefaultRoutingPolicyRequestFieldPathsTopLevel = []string{
	"downlink",
	"uplink",
}
var ListForwarderRoutingPoliciesRequestFieldPathsNested = []string{
	"limit",
	"page",
}

var ListForwarderRoutingPoliciesRequestFieldPathsTopLevel = []string{
	"limit",
	"page",
}
var PacketBrokerRoutingPoliciesFieldPathsNested = []string{
	"policies",
}

var PacketBrokerRoutingPoliciesFieldPathsTopLevel = []string{
	"policies",
}
var SetPacketBrokerRoutingPolicyRequestFieldPathsNested = []string{
	"downlink",
	"downlink.application_data",
	"downlink.join_accept",
	"downlink.mac_data",
	"home_network_id",
	"home_network_id.net_id",
	"home_network_id.tenant_id",
	"uplink",
	"uplink.application_data",
	"uplink.join_request",
	"uplink.localization",
	"uplink.mac_data",
	"uplink.signal_quality",
}

var SetPacketBrokerRoutingPolicyRequestFieldPathsTopLevel = []string{
	"downlink",
	"home_network_id",
	"uplink",
}
var ListHomeNetworksRequestFieldPathsNested = []string{
	"limit",
	"page",
}

var ListHomeNetworksRequestFieldPathsTopLevel = []string{
	"limit",
	"page",
}
var ListHomeNetworksRoutingPoliciesRequestFieldPathsNested = []string{
	"home_network_id",
	"home_network_id.net_id",
	"home_network_id.tenant_id",
	"limit",
	"offset",
}

var ListHomeNetworksRoutingPoliciesRequestFieldPathsTopLevel = []string{
	"home_network_id",
	"limit",
	"offset",
}
