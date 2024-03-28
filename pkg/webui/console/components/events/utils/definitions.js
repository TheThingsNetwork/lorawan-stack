// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import {
  IconEventClearAll,
  IconEventConnection,
  IconEventCreate,
  IconEventDelete,
  IconEventDownlink,
  IconEventError,
  IconEventGatewayConnect,
  IconEventGatewayDisconnect,
  IconEventJoin,
  IconEventMode,
  IconEventRekey,
  IconEventStatus,
  IconEventSwitch,
  IconEventUpdate,
  IconEventUplink,
} from '@ttn-lw/components/icon'

import DownlinkMessage from '../previews/downlink-message'
import GatewayUplinkMessage from '../previews/gateway-uplink-message'
import ApplicationDownlink from '../previews/application-downlink'
import ApplicationUplink from '../previews/application-uplink'
import ApplicationUplinkNormalized from '../previews/application-uplink-normalized'
import ApplicationUp from '../previews/application-up'
import ApplicationLocation from '../previews/application-location'
import JoinRequest from '../previews/join-request'
import JoinResponse from '../previews/join-response'
import GatewayStatus from '../previews/gateway-status'
import ErrorDetails from '../previews/error-details'
import Value from '../previews/value'

export const eventIconMap = [
  {
    test: /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+create$/,
    icon: IconEventCreate,
  },
  {
    test: /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+update$/,
    icon: IconEventUpdate,
  },
  {
    test: /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+delete$/,
    icon: IconEventDelete,
  },
  {
    test: /^(ns|as)\.up(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/,
    icon: IconEventUplink,
  },
  {
    test: /^(ns|as)\.down(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/,
    icon: IconEventDownlink,
  },
  {
    test: /^(js|ns|as)(\.up|\.down)?\.(join|rejoin)(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/,
    icon: IconEventJoin,
  },
  {
    test: /^gs\.up(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/,
    icon: IconEventUplink,
  },
  {
    test: /^gs\.down(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/,
    icon: IconEventDownlink,
  },
  {
    test: /^gs.gateway.connect$/,
    icon: IconEventGatewayConnect,
  },
  {
    test: /^gs.gateway.disconnect$/,
    icon: IconEventGatewayDisconnect,
  },
  {
    test: /^ns\.mac\.rekey\..*$/,
    icon: IconEventRekey,
  },
  {
    test: /^ns\.mac\.device_mode\..*$/,
    icon: IconEventMode,
  },
  {
    test: /^ns\..*\.switch\..*$/,
    icon: IconEventSwitch,
  },
  {
    test: /^ns(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/,
    icon: IconEventConnection,
  },
  {
    test: /^gs\.status\..*$/,
    icon: IconEventStatus,
  },
  {
    test: /^synthetic\.status\.cleared$/,
    icon: IconEventClearAll,
  },
  {
    test: /^synthetic\.error\..*$/,
    icon: IconEventError,
  },
]

export const dataTypeMap = {
  ApplicationDownlink,
  ApplicationUplink,
  ApplicationUplinkNormalized,
  ApplicationUp,
  ApplicationLocation,
  DownlinkMessage,
  GatewayUplinkMessage,
  JoinRequest,
  JoinResponse,
  ErrorDetails,
  GatewayStatus,
  Value,
}

export const applicationUpMessages = [
  'uplink_message',
  'uplink_normalized',
  'join_accept',
  'downlink_ack',
  'downlink_nack',
  'downlink_sent',
  'downlink_failed',
  'downlink_queued',
  'downlink_queue_invalidated',
  'location_solved',
  'service_data',
]
