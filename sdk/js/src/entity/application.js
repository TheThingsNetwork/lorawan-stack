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

import Devices from '../service/devices'
import Entity from './entity'

/**
 * Application Class wraps the single application data and provides abstractions
 * that simplify communication with the API.
 * @extends Entity
 */
class Application extends Entity {
  constructor (parent, rawData, isNew = true) {
    let data = rawData
    if ('application' in rawData) {
      data = rawData.application
    }

    super(data, isNew)

    this._parent = parent
    this._id = data.ids.application_id
    this.Devices = new Devices(parent._api, { stackConfig: parent._stackConfig })
  }

  // API Keys

  async getApiKeys (params) {
    return this._parent.ApiKeys.getAll(this._id, params)
  }

  async addApiKey (key) {
    return this._parent.ApiKeys.getAll(this._id, key)
  }

  // Devices

  async getDevice (deviceId) {
    return this._parent.Devices.getById(this._id, deviceId)
  }

  async getDevices (params) {
    return this._parent.Devices.getAll(this._id, params)
  }

  async save (userId) {
    let res

    if (this._isNew) {
      res = await this._parent.create(userId, this.toObject())
    } else {
      const updateMask = super.getUpdateMask()
      res = await this._parent.updateById(this._id, super.mask(updateMask))
    }
    super.save(res)

    return res
  }
}

export default Application
