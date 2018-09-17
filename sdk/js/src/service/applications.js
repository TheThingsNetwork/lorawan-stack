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
    const applications = await this._api.ListApplications()
    return applications
  }

  async getById (id) {
    const application = await this._api.GetApplication({ 'application_ids.application_id': id })
    return new Application(this, application, false)
  }

  async getByOrganization (organizationId) {
    return this._api.ListApplications({ 'collaborator.organization_ids.organization_id': organizationId })
  }

  async getByCollaborator (userId) {
    return this._api.ListApplications({ 'collaborator.user_ids.user_id': userId })
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    return this._api.UpdateApplication({
      'application.ids.application_id': id,
    },
    {
      application: patch,
      field_mask: Marshaler.fieldMask(mask),
    })
  }

  // Create

  async create (userId = this._defaultUserId, application) {
    return this._api.CreateApplication({ 'collaborator.user_ids.user_id': userId }, { application })
  }

  // Delete

  async deleteById (applicationId) {
    return this._api.DeleteApplication({ application_id: applicationId })
  }

  // Shorthand to methods of single application
  withId (id) {
    const parent = this
    const api = parent._api
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

export default Applications
