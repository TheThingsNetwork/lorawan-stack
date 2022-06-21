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

import { isPlainObject } from 'lodash'

import { ACTIVATION_MODES } from '@console/lib/device-utils'

// End device selectors.

export const getActivationMode = device =>
  device.supports_join === true
    ? ACTIVATION_MODES.OTAA
    : device.multicast === true
    ? ACTIVATION_MODES.MULTICAST
    : device.supports_join === false && device.multicast === false
    ? ACTIVATION_MODES.ABP
    : ACTIVATION_MODES.NONE

export const getLorawanVersion = device => device.lorawan_version || '1.1.0'

export const getApplicationServerAddress = device => device.application_server_address
export const getNetworkServerAddress = device => device.network_server_address
export const getJoinServerAddress = device => device.join_server_address

// End device repository utils.

export const SELECT_OTHER_OPTION = '_other_'
export const isOtherOption = option => option === SELECT_OTHER_OPTION

/*
  `hardware_version` is not required when registering an end device in the device repository, so for 
  certain end device models it can be missing. When this is the case, we still want to allow the users
  to select such models because `firmware_version` (that might depend on hw version) and `band_id`
  are required. `SELECT_UNKNOWN_HW_OPTION` option represents such end devices.
*/
export const SELECT_UNKNOWN_HW_OPTION = '_unknown_hw_version_'
export const isUnknownHwVersion = option => option === SELECT_UNKNOWN_HW_OPTION

export const REGISTRATION_TYPES = Object.freeze({
  SINGLE: 'single',
  MULTIPLE: 'multiple',
})

// Misc utilities.

const recursiveRemove = (dest, src) => {
  for (const key in src) {
    if (isPlainObject(dest[key])) {
      return recursiveRemove(dest[key], src[key])
    }
    if (key in dest) {
      delete dest[key]
    }
  }
}

/** .
 * Recursively removes all common props of `dest` and `src` from `src`.
 * This is useful e.g. To remove initial props of a form section from the form values when it unmounts.
 *
 * @param {object} dest - The destination object from which the props are removed.
 * @param {object} src - The source object with the props that will be removed from the `dest`.
 * @returns {object} A new `dest` object that has all props of `src` deleted.
 */
export const subtractObject = (dest, src) => {
  const res = { ...dest }
  recursiveRemove(res, src)
  console.log(JSON.stringify(dest), JSON.stringify(src), JSON.stringify(res))

  return res
}
