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

import Marshaler from '../util/marshaler'
import Device from '../entity/device'

/**
 * Devices Class provides an abstraction on all devices and manages data
 * handling from different sources. It exposes an API to easily work with
 * device data.
 */
class Devices {
  constructor (api, { applicationId, proxy = true }) {
    this._api = api
    this._applicationId = applicationId
    this._idMask = { route: { 'end_device.ids.application_ids.application_id': this._applicationId }}
    this._entityTransform = proxy
      ? app => new Device(this, app, false)
      : undefined
  }

  async getById (deviceId) {
    const res = await this._api.EndDeviceRegistry.Get({
      ...this._idMask,
      device_id: deviceId,
    })

    return new Device(res, this)
  }

  async updateById (deviceId) {
    return this._api.EndDeviceRegistry.Get({
      ...this._idMask,
      device_id: deviceId,
    })
  }

  async create (device, applicationId = this._applicationId) {
    const result = await this._api.EndDeviceRegistry.Create(
      { route: { 'end_device.ids.application_ids.application_id': applicationId }},
      { end_device: device }
    )
    return Marshaler.unwrapDevice(
      result,
      this._entityTransform
    )
  }
}

export default Devices
