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

import Marshaler from '../util/marshaler'
import { Devices, Device } from './devices'
import Entity from './entity'

/**
 * Applications Class provides an abstraction on all applications and manages
 * data handling from different sources. It exposes an API to easily work with
 * application data.
 */
class Applications {
  constructor (api, { defaultUserId }) {
    this.defaultUserId = defaultUserId
    this.api = api
  }

  // Retrieval

  async getAll () {
    const applications = await this.api.ListApplications()
    return applications
  }

  async getById (id) {
    const data = await this.api.GetApplication({ 'application_ids.application_id': id })
    return new Application(this, data, false)
  }

  async getByOrganization (organizationId) {
    return this.api.ListApplications({ 'collaborator.organization_ids.organization_id': organizationId })
  }

  async getByCollaborator (userId) {
    return this.api.ListApplications({ 'collaborator.user_ids.user_id': userId })
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    return this.api.UpdateApplication({
      'application.ids.application_id': id,
    },
    {
      application: patch,
      field_mask: Marshaler.fieldMask(mask),
    })
  }

  // Create

  async create (userId = this.defaultUserId, application) {
    return this.api.CreateApplication({ 'collaborator.user_ids.user_id': userId }, { application })
  }

  // Delete

  async deleteById (applicationId) {
    return this.api.DeleteApplication({ application_id: applicationId })
  }

  // Shorthand to methods of single application
  withId (id) {
    const parent = this
    const api = parent.api
    const idMask = { 'application_ids.application_id': id }
    return {
      async getDevices () {
        return api.ListDevices(idMask)
      },
      async getDevice (deviceId) {
        const result = await api.GetDevice({ ...idMask, device_id: deviceId })
        return new Device(result, api)
      },
      async getApiKeys () {
        return api.ListApplicationAPIKeys(idMask)
      },
      async getCollaborators () {
        return api.ListApplicationCollaborators(idMask)
      },
      async addApiKey (key) {
        return api.GenerateApplicationAPIKey(idMask, key)
      },
      async addCollaborator (collaborator) {
        return api.SetApplicationCollaborator(idMask, collaborator)
      },
      async updateDevice (device) {
        return api.GetDevice(idMask, device)
      },
    }
  }
}

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
    this.Devices = new Devices(parent.api, this._id)
  }

  // Collaborators

  async getCollaborators () {
    return this._parent.withId(this._id).getCollaborators()
  }

  async addCollaborator (collaborator) {
    return this._parent.withId(this._id).setCollaborator(collaborator)
  }

  // API Keys

  async getApiKeys () {
    return this._parent.withId(this._id).getApiKeys()
  }

  async addApiKey (key) {
    return this._parent.withId(this._id).addApiKey(key)
  }

  // Devices

  async getDevice (deviceId) {
    return this._parent.withId(this._id).getDevice(deviceId)
  }

  async getDevices () {
    return this._parent.withId(this._id).getDevices()
  }

  async save (userId) {
    let res

    try {
      if (this._isNew) {
        res = await this._parent.create(userId, this.toObject())
      } else {
        const updateMask = super.getUpdateMask()
        res = await this._parent.updateById(this._id, super.mask(updateMask))
      }
      super.save(res)

      return res
    } catch (err) {
      throw err
    }
  }
}

export { Applications as default, Application, Applications }
