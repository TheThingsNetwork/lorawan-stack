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
 * @param {Object} api - The connector to be used by the service.
 * @param {Object} config - The configuration for the service
 * @param {string} config.defaultUserId - The users identifier to be used in
 * user related requests.
 * @param {boolean} config.proxy - The flag to identify if the results
 *  should be proxied with the wrapper objects.
 */
class Applications {
  constructor (api, { defaultUserId, proxy = true }) {
    this._defaultUserId = defaultUserId
    this._api = api
    this._applicationTransform = proxy
      ? app => new Application(this, app, false)
      : undefined

    this.getAll = this.getAll.bind(this)
  }

  // Retrieval

  async getAll (params) {
    const result = await this._api.ApplicationRegistry.List({ query: params })
    return Marshaler.unwrapApplications(
      result,
      this._applicationTransform
    )
  }

  async getById (id) {
    const result = await this._api.ApplicationRegistry.Get({
      route: { 'application_ids.application_id': id },
    })

    return Marshaler.unwrapApplication(
      result,
      this._applicationTransform
    )
  }

  async getByOrganization (organizationId) {
    return this._api.ApplicationRegistry.List({
      route: { 'collaborator.organization_ids.organization_id': organizationId },
    })
  }

  async getByCollaborator (userId) {
    return this._api.ApplicationRegistry.List({
      route: { 'collaborator.user_ids.user_id': userId },
    })
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    return this._api.ApplicationRegistry.Update({
      route: {
        'application.ids.application_id': id,
      },
    },
    {
      application: patch,
      field_mask: Marshaler.fieldMask(mask),
    })
  }

  // Create

  async create (userId = this._defaultUserId, application) {
    return this._api.ApplicationRegistry.Create({
      route: { 'collaborator.user_ids.user_id': userId },
    },
    { application })
  }

  // Delete

  async deleteById (applicationId) {
    return this._api.ApplicationRegistry.Delete({
      route: { application_id: applicationId },
    })
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
