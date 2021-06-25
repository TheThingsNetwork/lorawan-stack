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
    'Packet Broker can be used to exchange traffic with other LoRaWAN networks to share coverage and improve the overall network performance.',
  packetBrokerWebsite: 'Packet Broker website',
  registerNetwork: 'Register network',
  registerThisNetwork: 'Register this network',
  network: 'Network: {network}',
  homeNetworkEnabled: 'Home network enabled',
  homeNetworkDisabled: 'Home network disabled',
  forwarderEnabled: 'Forwarder enabled',
  forwarderDisabled: 'Forwarder disabled',
  packetBrokerRegistrationDesc:
    "To enable package exchange from or to your home network, it is necessary to register your network. This will make your network known to the Packet Broker and enable you to configure your network's package exchange behavior.",
  packetBrokerDisabledDesc:
    'It appears like The Things Stack is currently not set up to use Packet Broker. Please refer to the documentation link above for more information about how to set up The Things Stack for peering via Packet Broker.',
  routingPolicyInformation:
    'You can use the checkboxes below to control the default forwarding behavior of your network. You can additionally set up individual per-network routing policies via the Network tab.',
  defaultRoutingPolicySet: 'Default routing policy set',
  routingPolicySet: 'Routing policy set',
  defaultRoutingPolicy: 'Default routing policy',
  networks: 'Networks',
  networkInformation: 'Network information',
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
  deregisterModal: `Are you sure you want to deregister your network from the Packet Broker?{lineBreak}This will <b>instantly disable any package exchange</b> with foreign networks.`,
})
