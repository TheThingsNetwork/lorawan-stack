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

import autoBind from 'auto-bind'

import hexToBase64 from '../util/bytes'
import Marshaler from '../util/marshaler'

class DeviceClaim {
  constructor(registry, { stackConfig }) {
    this._api = registry
    this._stackConfig = stackConfig
    autoBind(this)
  }

  // Claim
  async claim(applicationId, qrCode, values) {
    const deviceToClaim = qrCode ? { qr_code: qrCode } : values
    const payload = {
      ...deviceToClaim,
      target_application_ids: {
        application_id: applicationId,
      },
    }

    const response = await this._api.EndDeviceClaimingServer.Claim(undefined, payload)

    return Marshaler.payloadSingleResponse(response)
  }

  async GetInfoByJoinEUI(join_eui) {
    const response = await this._api.EndDeviceClaimingServer.GetInfoByJoinEUI(undefined, join_eui)

    return Marshaler.payloadSingleResponse(response)
  }

  /**
   * Unclaim an end device.
   *
   * @param {string} applicationId - The Application ID.
   * @param {string} deviceId - The Device ID.
   * @param {Array} devEui - The Device dev_eui.
   * @param {object} joinEui - The Device join_eui.
   * @returns {object} - An empty object on successful requests, an error otherwise.
   */
  async unclaim(applicationId, deviceId, devEui, joinEui) {
    const params = {
      routeParams: {
        'application_ids.application_id': applicationId,
        device_id: deviceId,
      },
    }

    const response = await this._api.EndDeviceClaimingServer.Unclaim(params, {
      dev_eui: hexToBase64(devEui),
      join_eui: hexToBase64(joinEui),
    })

    return Marshaler.payloadSingleResponse(response)
  }
}

export default DeviceClaim
