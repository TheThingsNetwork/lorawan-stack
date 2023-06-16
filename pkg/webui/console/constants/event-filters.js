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

import { APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

export const END_DEVICE_EVENTS_VERBOSE_FILTERS = [
  'as.*.drop',
  'as.down.data.receive',
  'as.up.*.forward',
  'js.join.accept',
  'js.join.reject',
  'ns.up.data.process',
  'ns.up.join.process',
  'ns.down.data.schedule.attempt',
  'ns.mac.*.answer.reject',
  '*.warning',
  '*.fail',
  'end_device.*',
]

export const APPLICATION_EVENTS_VERBOSE_FILTERS = [
  ...END_DEVICE_EVENTS_VERBOSE_FILTERS,
  'application.*',
]

export const GATEWAY_EVENTS_VERBOSE_FILTERS = [
  'gs.down.send',
  'gs.gateway.connect',
  'gs.gateway.disconnect',
  'gs.status.receive',
  'gs.up.receive',
  '*.warning',
  '*.fail',
  'gateway.*',
]

// Regex for matching heartbeat events that trigger an update of the
// last activity display.
export const EVENT_END_DEVICE_HEARTBEAT_FILTERS_REGEXP = /^as.up\..*\.forward$/

// Utility function to convert filter arrays to Regular Expressions strings
// that the backend accepts for applying filters.
const filterListToRegExpList = array =>
  array.map(f => `/^${f.replace(/\./g, '\\.').replace(/\*/g, '.*')}$/`)

export const EVENT_FILTERS = {
  [APPLICATION]: [
    {
      id: 'default',
      filter: APPLICATION_EVENTS_VERBOSE_FILTERS,
      filterRegExp: filterListToRegExpList(APPLICATION_EVENTS_VERBOSE_FILTERS),
    },
  ],
  [END_DEVICE]: [
    {
      id: 'default',
      filter: END_DEVICE_EVENTS_VERBOSE_FILTERS,
      filterRegExp: filterListToRegExpList(END_DEVICE_EVENTS_VERBOSE_FILTERS),
    },
  ],
  [GATEWAY]: [
    {
      id: 'default',
      filter: GATEWAY_EVENTS_VERBOSE_FILTERS,
      filterRegExp: filterListToRegExpList(GATEWAY_EVENTS_VERBOSE_FILTERS),
    },
  ],
}
