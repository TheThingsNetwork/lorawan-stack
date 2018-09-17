// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import Entity from './entity'

/**
 * Devices Class provides an abstraction on all devices and manages data
 * handling from different sources. It exposes an API to easily work with
 * device data.
 */
class Devices {
  constructor (api, applicationId) {
    this.api = api
    this.applicationId = applicationId
    this.idMask = { 'application_ids.application_id': this.applicationId }
  }

  async getById (deviceId) {
    const res = await this.api.GetDevice({
      ...this.idMask,
      device_id: deviceId,
    })

    return new Device(res, this)
  }

  async updateById (deviceId) {
    return this.api.GetDevice(this.idMask)
  }
}

/**
 * Devices Class wraps the single device data and provides abstractions that
 * simplify communication with the API.
 * @extends Entity
 */
class Device extends Entity {
  constructor (data, api) {
    super(data)

    // TODO: Check for data validity

    this._deviceId = data.ids.device_id
    this._appId = data.ids.application_ids.application_id
    this._appIdMask = { 'application_ids.application_id': this._appId }
    this._api = api
  }

  save () {
    return this.api.SetDevice({ ...this._appIdMask, device_id: this._deviceId }, this.toObject())
  }
}

export { Devices as default, Devices, Device }
