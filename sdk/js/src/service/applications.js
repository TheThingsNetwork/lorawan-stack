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
import Devices from '../service/devices'
import Application from '../entity/application'
import ApiKeys from './api-keys'
import Link from './link'
import Collaborators from './collaborators'

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
  constructor (api, { defaultUserId, stackConfig, proxy = true }) {
    this._defaultUserId = defaultUserId
    this._api = api
    this._proxy = proxy

    this.ApiKeys = new ApiKeys(api.ApplicationAccess, {
      parentRoutes: {
        list: 'application_id',
        create: 'application_ids.application_id',
        update: 'application_ids.application_id',
      },
    })
    this.Link = new Link(api.As)
    this.Devices = new Devices(api, { proxy, stackConfig })
    this.Collaborators = new Collaborators(api.ApplicationAccess, {
      parentRoutes: {
        list: 'application_id',
        set: 'application_ids.application_id',
      },
    })
  }

  _responseTransform (response, single = true) {
    return Marshaler[single ? 'unwrapApplication' : 'unwrapApplications'](
      response,
      this._proxy
        ? app => new Application(this, app, false)
        : undefined
    )
  }

  // Retrieval

  async getAll (params) {
    const response = await this._api.ApplicationRegistry.List(undefined, params)

    return this._responseTransform(response, false)
  }

  async getById (id, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.ApplicationRegistry.Get({
      routeParams: { 'application_ids.application_id': id },
    }, fieldMask)

    return this._responseTransform(response)
  }

  async getByOrganization (organizationId) {
    const response = this._api.ApplicationRegistry.List({
      routeParams: { 'collaborator.organization_ids.organization_id': organizationId },
    })

    return this._responseTransform(response)
  }

  async getByCollaborator (userId) {
    const response = this._api.ApplicationRegistry.List({
      routeParams: { 'collaborator.user_ids.user_id': userId },
    })

    return this._responseTransform(response)
  }

  async search (params) {
    const response = await this._api.EntityRegistrySearch.SearchApplications({
      queryParams: params,
    })

    return this._responseTransform(response, false)
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    const response = await this._api.ApplicationRegistry.Update({
      routeParams: {
        'application.ids.application_id': id,
      },
    },
    {
      application: patch,
      field_mask: Marshaler.fieldMask(mask),
    })
    return Marshaler.unwrapApplication(
      response,
      this._applicationTransform
    )
  }

  // Create

  async create (userId = this._defaultUserId, application) {
    const response = await this._api.ApplicationRegistry.Create({
      routeParams: { 'collaborator.user_ids.user_id': userId },
    },
    { application })
    return this._responseTransform(response)
  }

  // Delete

  async deleteById (applicationId) {
    return this._api.ApplicationRegistry.Delete({
      routeParams: { application_id: applicationId },
    })
  }

  async getRightsById (applicationId) {
    const result = await this._api.ApplicationAccess.ListRights({
      routeParams: { application_id: applicationId },
    })

    return Marshaler.unwrapRights(result)
  }
}

export default Applications
