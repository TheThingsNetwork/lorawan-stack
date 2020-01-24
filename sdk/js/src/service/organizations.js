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
import ApiKeys from './api-keys'
import Collaborators from './collaborators'

class Organizations {
  constructor(api) {
    this._api = api

    this.ApiKeys = new ApiKeys(api.OrganizationAccess, {
      parentRoutes: {
        get: 'organization_ids.organization_id',
        list: 'organization_ids.organization_id',
        create: 'organization_ids.organization_id',
        update: 'organization_ids.organization_id',
      },
    })
    this.Collaborators = new Collaborators(api.OrganizationAccess, {
      parentRoutes: {
        get: 'organization_ids.organization_id',
        list: 'organization_ids.organization_id',
        set: 'organization_ids.organization_id',
      },
    })
  }

  // Retrieval

  async getAll(params, selector) {
    const response = await this._api.OrganizationRegistry.List(undefined, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return Marshaler.payloadListResponse('organizations', response)
  }

  async getById(id, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.OrganizationRegistry.Get(
      {
        routeParams: { 'organization_ids.organization_id': id },
      },
      fieldMask,
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async search(params, selector) {
    const response = await this._api.EntityRegistrySearch.SearchOrganizations(undefined, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return Marshaler.payloadListResponse('organizations', response)
  }

  // Create

  async create(userId, organization) {
    const response = await this._api.OrganizationRegistry.Create(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId },
      },
      { organization },
    )

    return Marshaler.payloadSingleResponse(response)
  }

  // Update

  async updateById(id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    const response = await this._api.OrganizationRegistry.Update(
      {
        routeParams: {
          'organization.ids.organization_id': id,
        },
      },
      {
        organization: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )

    return Marshaler.payloadSingleResponse(response)
  }

  // Delete

  async deleteById(organizationId) {
    const response = await this._api.OrganizationRegistry.Delete({
      routeParams: { organization_id: organizationId },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async getRightsById(organizationId) {
    const result = await this._api.OrganizationAccess.ListRights({
      routeParams: { organization_id: organizationId },
    })

    return Marshaler.unwrapRights(result)
  }

  // Events Stream

  async openStream(identifiers, tail, after) {
    const payload = {
      identifiers: identifiers.map(id => ({
        organization_ids: { organization_id: id },
      })),
      tail,
      after,
    }

    return this._api.Events.Stream(undefined, payload)
  }
}

export default Organizations
