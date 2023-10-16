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

import { defineMessages } from 'react-intl'

export default defineMessages({
  packetBrokerInfoText:
    'Packet Broker can be used to exchange traffic (peering) with other LoRaWAN networks to share coverage and improve the overall network performance.',
  packetBrokerWebsite: 'Packet Broker website',
  registrationStatus: 'Registration status',
  registerNetwork: 'Register network',
  networkVisibility: 'Network visibility',
  packetBrokerRegistrationDesc:
    "To enable peering from or to your home network, it is necessary to register your network. This will make your network known to Packet Broker and enable you to configure your network's peering behavior.",
  packetBrokerDisabledDesc:
    'The Things Stack is not set up to use Packet Broker. Please refer to the documentation link above for instructions on how to set up The Things Stack for peering with Packet Broker.',
  packetBrokerRegistrationDisabledDesc:
    'The Things Stack is set up to use Packet Broker, but security settings disallow (de)registering your network here. Please contact Packet Broker to manage your registration. Refer to the documentation link above for contact information.',
  network: 'Network: {network}',
  homeNetworkEnabled: 'Home network <b>enabled</b>',
  homeNetworkDisabled: 'Home network <b>disabled</b>',
  forwarderEnabled: 'Forwarder <b>enabled</b>',
  forwarderDisabled: 'Forwarder <b>disabled</b>',
  listNetwork: 'List network publicly',
  listNetworkDesc:
    'Listing your network allows other network administrators to see your network. This allows them to easily configure routing policies with your network.',
  unlistNetwork: 'Unlist this network',
  confirmUnlist: 'Confirm unlist',
  unlistModal:
    'Are you sure you want to unlist your network in Packet Broker?{lineBreak}' +
    'This will hide your network. Other network administrators will not be able to see your network to configure routing policies.',
  routingPolicyInformation:
    'You can use the checkboxes below to control the default forwarding behavior of your network. You can additionally set up individual per-network routing policies via the Networks tab.',
  defaultRoutingPolicySet: 'Default routing policy set',
  routingPolicySet: 'Routing policy set',
  defaultRoutingPolicy: 'Default routing policy',
  devAddressBlock: 'Device address block',
  devAddressBlocks: 'Device address blocks',
  lastPolicyChange: 'Last policy change',
  networkId: 'Network ID',
  routingPolicyFromThisNetwork: "This network's routing policy towards us",
  routingPolicyToThisNetwork: 'Set routing policy towards this network',
  saveRoutingPolicy: 'Save routing policy',
  noPolicySet: 'No policy set yet',
  prefixes: 'Prefixes',
  homeNetworkClusterId: 'Home network cluster ID',
  backToAllNetworks: 'Back to all networks',
  deregisterNetwork: 'Deregister this network',
  confirmDeregister: 'Confirm deregistration',
  deregisterModal:
    'Are you sure you want to deregister your network from Packet Broker?{lineBreak}' +
    'This will <b>permanently delete</b> all routing policies and may stop traffic from flowing.{lineBreak}' +
    'Traffic may still be forwarded to your network based on default routing policies configured by forwarders.',
  defaultGatewayVisibility: 'Default gateway visibility',
  gatewayVisibilityInformation:
    'You can use the checkboxes to control what information of your gateways will be visible. Note that this information will be visible to the public and not only to registered networks.',
  defaultGatewayVisibilitySet: 'Default gateway visibility set',
  packetBrokerStatusPage: 'Packet Broker Status page',
})
