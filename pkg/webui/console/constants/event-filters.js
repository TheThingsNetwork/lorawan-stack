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

export const EVENT_VERBOSE_FILTERS = [
  'as.*.drop',
  'as.down.data.forward',
  'as.up.data.forward',
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

// A RegExp converted from the glob list of filtered event names.
export const EVENT_VERBOSE_FILTERS_REGEXP = EVENT_VERBOSE_FILTERS.reduce(
  (acc, cur, i) => `${acc}${i !== 0 ? '|' : ''}${cur.replace('.', '\\.').replace('*', '.*')}`,
  '',
)

// A map that allows to translate back the filter list from the converted
// RegExp string. Useful to show a human readable filter list in the event
// stream, which only uses the RegExp string internally.
export const EVENT_FILTER_MAP = Object.freeze({
  [EVENT_VERBOSE_FILTERS_REGEXP]: EVENT_VERBOSE_FILTERS,
})
