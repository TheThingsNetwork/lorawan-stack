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

import { selectJsConfig } from '@ttn-lw/lib/selectors/env'
import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'

/**
 * Returns whether the device is OTAA.
 * Note: device type is mainly derived based on the `supports_join` and
 * `multicast` fields.
 * However, in cases when NS is not available, `root_keys` can be used to
 * determine whether the device is OTAA.
 *
 * @param {object} device - The device object.
 * @returns {boolean} `true` if the device is OTAA, `false` otherwise.
 */
export const isDeviceOTAA = device =>
  Boolean(device) && (Boolean(device.supports_join) || Boolean(device.root_keys))

/**
 * Returns whether the device is ABP.
 *
 * @param {object} device - The device object.
 * @returns {boolean} `true` if the device is ABP, `false` otherwise.
 */
export const isDeviceABP = device =>
  Boolean(device) && !Boolean(device.supports_join) && !Boolean(device.multicast)

/**
 * Returns whether the device is multicast.
 *
 * @param {object} device - The device object.
 * @returns {boolean} `true` if the device is multicast, `false` otherwise.
 */
export const isDeviceMulticast = device => Boolean(device) && Boolean(device.multicast)

/**
 * Returns whether an end device is provisioned on an external join server.
 *
 * @param {object} device - The device object.
 * @returns {boolean} `true` if the end device is provisioned on an external
 * join server, `false` otherwise.
 */
export const hasExternalJs = device => {
  const { enabled, base_url } = selectJsConfig()

  const deviceJs = device.join_server_address
  const stackJs = getHostnameFromUrl(base_url)

  return !enabled || typeof deviceJs === 'undefined' || deviceJs !== stackJs
}

/**
 * Returns whether an end device has joined the network.
 *
 * @param {object} device - The device object.
 * @returns {boolean} `true` if the end device has join thr network, `false` otherwise.
 */
export const isDeviceJoined = device =>
  Boolean(device) &&
  Boolean(device.session) &&
  Boolean(device.session.dev_addr) &&
  Boolean(device.session.keys) &&
  Boolean(Object.keys(device.session.keys).length)
