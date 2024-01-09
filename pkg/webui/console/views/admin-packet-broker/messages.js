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
    'Packet Broker is a service by The Things Industries to facilitate peering between LoRaWAN networks. This extends network coverage and improves overall network performance and device battery lifetime.',
  packetBrokerWebsite: 'Packet Broker website',
  learnMore: 'Learn more',
  whyNetworkPeeringTitle: 'Why choose network peering?',
  whyNetworkPeeringText:
    'Since LoRaWAN uses shared spectrum, gateways receive messages from devices registered on other LoRaWAN networks. Instead of discarding this traffic, these messages can be forwarded via Packet Broker to the home network of these devices. This extends coverage of networks and allows devices to use higher data rates that reduce channel utilization and increase battery life. No sensitive data is exposed as LoRaWAN is end-to-end encrypted and integrity protected.',
  enbaling:
    'Enable forwarding via the options below or define custom routing policies. In the Networks tab below (visible by selecting the option "Use custom routing policies"), you can see which other networks are forwarding data to this network.',
  packetBrokerDisabledDesc:
    'The Things Stack is not set up to use Packet Broker. Please refer to the documentation link above for instructions on how to set up The Things Stack for peering with Packet Broker.',
  enablePacketBroker: 'Enable Packet Broker',
  packetBrokerRegistrationDesc:
    'Enabling will allow other networks to send traffic to you as well as you forwarding traffic to them, based on the exact routing policy.',
  routingConfig: 'Routing configuration',
  network: 'Network: {network}',
  listNetwork: 'List my network in Packet Broker publicly',
  listNetworkDesc:
    'Public listing will make it easier for other network operators to set up routing policies for your network. Hence public listing is generally recommended.',
  unlistNetwork: 'Unlist this network',
  confirmUnlist: 'Confirm unlist',
  unlistModal:
    'Are you sure you want to unlist your network in Packet Broker?{lineBreak}' +
    'This will hide your network. Other network administrators will not be able to see your network to configure routing policies.',
  routingPolicyInformation:
    'You can use the checkboxes below to control the default forwarding behavior of your network. You can additionally set up individual per-network routing policies via the Networks tab.',
  defaultRoutingPolicySet: 'Default routing configuration set',
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
