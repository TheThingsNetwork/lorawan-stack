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
import ApiKeys from './api-keys'
import Link from './link'

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

    this.ApiKeys = new ApiKeys(api.ApplicationAccess, {
      list: 'application_id',
      create: 'application_ids.application_id',
      update: 'application_ids.application_id',
    })
    this.Link = new Link(api.As)
<<<<<<< HEAD
=======
    this.Device = new Device(api, { proxy })

    this.getAll = this.getAll.bind(this)
    this.getById = this.getById.bind(this)
    this.getByOrganization = this.getByOrganization.bind(this)
    this.getByCollaborator = this.getByCollaborator.bind(this)
    this.search = this.search.bind(this)
    this.updateById = this.updateById.bind(this)
    this.create = this.create.bind(this)
    this.deleteById = this.deleteById.bind(this)
    this.getRightsById = this.getRightsById.bind(this)
>>>>>>> 34741e7f7... util: Remove withId shorthand
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
    const result = this._api.ApplicationRegistry.List({
      route: { 'collaborator.organization_ids.organization_id': organizationId },
    })

    return Marshaler.unwrapApplications(
      result,
      this._applicationTransform
    )
  }

  async getByCollaborator (userId) {
    const result = this._api.ApplicationRegistry.List({
      route: { 'collaborator.user_ids.user_id': userId },
    })

    return Marshaler.unwrapApplications(
      result,
      this._applicationTransform
    )
  }

  async search (params) {
    const result = await this._api.EntityRegistrySearch.SearchApplications({
      query: params,
    })
    return Marshaler.unwrapApplications(
      result,
      this._applicationTransform
    )
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    const result = await this._api.ApplicationRegistry.Update({
      route: {
        'application.ids.application_id': id,
      },
    },
    {
      application: patch,
      field_mask: Marshaler.fieldMask(mask),
    })
    return Marshaler.unwrapApplication(
      result,
      this._applicationTransform
    )
  }

  // Create

  async create (userId = this._defaultUserId, application) {
    const result = await this._api.ApplicationRegistry.Create({
      route: { 'collaborator.user_ids.user_id': userId },
    },
    { application })
    return Marshaler.unwrapApplication(
      result,
      this._applicationTransform
    )
  }

  // Delete

  async deleteById (applicationId) {
    return this._api.ApplicationRegistry.Delete({
      route: { application_id: applicationId },
    })
  }

  async getRightsById (applicationId) {
    const result = await this._api.ApplicationAccess.ListRights({
      route: { application_id: applicationId },
    })

    return Marshaler.unwrapRights(result)
  }
}

export default Applications
