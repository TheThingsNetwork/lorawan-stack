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
  applicationDataAllowDesc: 'Allow downlink messages with FPort between 1 and 255',
  applicationDataDesc: 'Forward uplink messages with FPort 1-255',
  joinAcceptDesc: 'Allow join accept messages',
  joinRequest: 'Join request',
  joinRequestDesc: 'Forward join-request messages',
  localizationInformation: 'Localization  information',
  localizationInformationDesc: 'Forward gateway location, RSSI, SNR and fine timestamp',
  macDataAllowDesc: 'Allow downlink messages with FPort of 0',
  macDataDesc: 'Forward uplink messages with FPort 0',
  signalQualityInformation: 'Signal quality information',
  signalQualityInformationDesc: 'Forward RSSI and SNR',
  forwardsJoinRequest: 'Join request messages are forwarded',
  doesNotForwardJoinRequest: 'Join request messages are not forwarded',
  forwardsMacData: 'MAC data is forwarded',
  doesNotForwardMacData: 'MAC data is not forwarded',
  forwardsApplicationData: 'Application data is forwarded',
  doesNotForwardApplicationData: 'Application data is not forwarded',
  forwardsSignalQuality: 'Signal quality information is forwarded',
  doesNotForwardSignalQuality: 'Signal quality information is not forwarded',
  forwardsLocalization: 'Localization information is forwarded',
  doesNotForwardLocalization: 'Localization information is not forwarded',
  allowsJoinAccept: 'Join accept messages are allowed',
  doesNotAllowJoinAccept: 'Join accept messages are not allowed',
  allowsMacData: 'MAC data is allowed',
  doesNotAllowMacData: 'MAC data is not allowed',
  allowsApplicationData: 'Application data is allowed',
  doesNotAllowApplicationData: 'Application data is not allowed',
  uplinkPolicies: 'This top row shows the uplink forwarding policies of this network',
  downlinkPolicies: 'This bottom row shows the downlink policies of this network',
  gatewayAntennaPlacementLabel: 'Antenna placement',
  gatewayAntennaPlacementDescription: 'Show antenna placement (indoor/outdoor)',
  gatewayAntennaCountLabel: 'Antenna count',
  gatewayFineTimestampsLabel: 'Fine timestamps',
  gatewayFineTimestampsDescription: 'Whether the gateway produces fine timestamps',
  gatewayContactInfoDescription: 'Show means to contact the gateway owner or operator',
  gatewayStatusDescription: 'Show whether the gateway is online or offline',
  gatewayPacketRatesLabel: 'Packet rates',
  gatewayPacketRatesDescription: 'Receive and transmission packet rates',
})
