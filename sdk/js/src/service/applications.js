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
import Application from '../entity/application'

/**
 * Applications Class provides an abstraction on all applications and manages
 * data handling from different sources. It exposes an API to easily work with
 * application data.
 */
class Applications {
  constructor (api, { defaultUserId }) {
    this._defaultUserId = defaultUserId
    this._api = api
  }

  // Retrieval

  async getAll () {
    let applications = await this._api.ApplicationRegistry.List()
    applications = Marshaler.unwrapApplications(applications)
    return applications.map(data => new Application(this, data, false))
  }

  async getById (id) {
    const application = await this._api.ApplicationRegistry.Get({ 'application_ids.application_id': id })
    return new Application(this, application, false)
  }

  async getByOrganization (organizationId) {
    return this._api.ApplicationRegistry.List({ 'collaborator.organization_ids.organization_id': organizationId })
  }

  async getByCollaborator (userId) {
    return this._api.ApplicationRegistry.List({ 'collaborator.user_ids.user_id': userId })
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    return this._api.ApplicationRegistry.Update({
      'application.ids.application_id': id,
    },
    {
      application: patch,
      field_mask: Marshaler.fieldMask(mask),
    })
  }

  // Create

  async create (userId = this._defaultUserId, application) {
    return this._api.ApplicationRegistry.Create({ 'collaborator.user_ids.user_id': userId }, { application })
  }

  // Delete

  async deleteById (applicationId) {
    return this._api.ApplicationRegistry.Delete({ application_id: applicationId })
  }

  // Shorthand to methods of single application
  withId (id) {
    const parent = this
    const api = parent._api
    const idMask = { 'application_ids.application_id': id }
    return {
      async getDevices () {
        return api.EndDeviceRegistry.List(idMask)
      },
      async getDevice (deviceId) {
        const result = await api.EndDeviceRegistry.Get({
          'end_device_ids.application_ids.application_id': id,
          device_id: deviceId,
        })
        return new Device(result, api)
      },
      async getApiKeys () {
        return api.ApplicationAccess.ListAPIKeys({ application_id: id })
      },
      async getCollaborators () {
        return api.ApplicationAccess.ListCollaborators({ application_id: id })
      },
      async addApiKey (key) {
        return api.ApplicationAccess.CreateAPIKey(idMask, key)
      },
      async addCollaborator (collaborator) {
        return api.ApplicationAccess.SetCollaborator(idMask, collaborator)
      },
    }
  }
}

export default Applications
