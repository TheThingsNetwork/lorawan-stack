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

// Utility function to convert filter arrays to Regular Expressions.
const filterListToRegExpString = array =>
  array.reduce(
    (acc, cur, i) => `${acc}${i !== 0 ? '|' : ''}${cur.replace('.', '\\.').replace('*', '.*')}`,
    '',
  )

export const EVENT_VERBOSE_FILTERS = [
  'as.*.drop',
  'as.down.data.forward',
  'as.up.location.forward',
  'as.up.data.forward',
  'as.up.service.forward',
  'gs.down.send',
  'gs.gateway.connect',
  'gs.gateway.disconnect',
  'gs.status.receive',
  'gs.up.receive',
  'js.join.accept',
  'js.join.reject',
  'ns.mac.*.answer.reject',
  '*.warning',
  '*.fail',
  'organization.*',
  'user.*',
  'gateway.*',
  'application.*',
  'end_device.*',
  'client.*',
  'oauth.*',
]

export const EVENT_END_DEVICE_HEARTBEAT_FILTERS = [
  'ns.up.data.receive',
  'ns.up.join.receive',
  'ns.up.rejoin.receive',
]

// Converted RegExps.
export const EVENT_VERBOSE_FILTERS_REGEXP = filterListToRegExpString(EVENT_VERBOSE_FILTERS)
export const EVENT_END_DEVICE_HEARTBEAT_FILTERS_REGEXP = filterListToRegExpString(
  EVENT_END_DEVICE_HEARTBEAT_FILTERS,
)

// A map that allows to translate back the filter list from the converted
// RegExp string. Useful to show a human readable filter list in the event
// stream, which only uses the RegExp string internally.
export const EVENT_FILTER_MAP = Object.freeze({
  [EVENT_VERBOSE_FILTERS_REGEXP]: EVENT_VERBOSE_FILTERS,
  [EVENT_END_DEVICE_HEARTBEAT_FILTERS_REGEXP]: EVENT_END_DEVICE_HEARTBEAT_FILTERS,
})
