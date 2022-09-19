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

export const STACK_COMPONENTS = ['as', 'is', 'ns', 'js', 'gs', 'edtc', 'qrg', 'gcs', 'dcs']

export const STACK_COMPONENTS_MAP = STACK_COMPONENTS.reduce((acc, curr) => {
  acc[curr] = curr
  return acc
}, {})

export const URI_PREFIX_STACK_COMPONENT_MAP = {
  as: 'as',
  ns: 'ns',
  js: 'js',
  gs: 'gs',
  edtc: 'edtc',
  qrg: 'qrg',
  gcs: 'gcs',
  edcs: 'dcs',
}

export const AUTHORIZATION_MODES = Object.freeze({
  KEY: 'key',
  SESSION: 'session',
})

export const RATE_LIMIT_RETRIES = 5
