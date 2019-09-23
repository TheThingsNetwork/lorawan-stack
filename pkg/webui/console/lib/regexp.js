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

// alphanumeric, dashes, lowercase
export const id = /^[a-z0-9]+(-[a-z0-9]+)*$/
export const apiKey = /^NNSXS.[A-Z0-9]{39}.[A-Z0-9]{52}$/
export const address = new RegExp(
  '^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$',
)
export const natsUrl = /^((\w+):)?(\/\/(([a-zA-z-0-9]+)?(:([a-zA-z-0-9]+))?@)?([^/?:]+)(:(\d+))?)?(\/?([^/?#][^?#]*)?)?(\?([^#]+))?(#(\w*))?/
export const mqttUrl = new RegExp('^(mqtts?)://[^\\s/$.?#].[^\\s]*$')
export const latitude = /^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?)$/
export const longitude = /^\s*[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$/
export const int32 = /^-?\d+$/
