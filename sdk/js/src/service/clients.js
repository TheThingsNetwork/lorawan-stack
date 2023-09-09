// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import autoBind from 'auto-bind'

import Marshaler from '../util/marshaler'

import Collaborators from './collaborators'

class Clients {
  constructor(registry) {
    this._api = registry

    this.Collaborators = new Collaborators(registry.ClientAccess, {
      parentRoutes: {
        get: 'client_ids.client_id',
        list: 'client_ids.client_id',
        set: 'client_ids.client_id',
        delete: 'client_ids.client_id',
      },
    })

    autoBind(this)
  }

  _addState(fieldMask, clients) {
    // Ensure to set STATE_REQUESTED if needed, which gets stripped as null
    // value from the backend response.
    if (fieldMask && fieldMask.field_mask.paths.includes('state') && !('state' in clients)) {
      clients.state = 'STATE_REQUESTED'
    }

    return clients
  }

  async getAll(params, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.ClientRegistry.List(undefined, {
      ...params,
      ...fieldMask,
    })

    const clients = Marshaler.payloadListResponse('clients', response)
    clients.clients.map(clients => this._addState(fieldMask, clients))

    return clients
  }

  async search(params, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.EntityRegistrySearch.SearchClients(undefined, {
      ...params,
      ...fieldMask,
    })

    const clients = Marshaler.payloadListResponse('clients', response)
    clients.clients.map(clients => this._addState(fieldMask, clients))

    return clients
  }

  async getById(id, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.ClientRegistry.Get(
      {
        routeParams: { 'client_ids.client_id': id },
      },
      fieldMask,
    )

    const client = this._addState(fieldMask, Marshaler.payloadSingleResponse(response))

    return client
  }

  async create(ownerId = this._defaultUserId, client, isUserOwner = true) {
    const routeParams = isUserOwner
      ? { 'collaborator.user_ids.user_id': ownerId }
      : { 'collaborator.organization_ids.organization_id': ownerId }
    const response = await this._api.ClientRegistry.Create(
      {
        routeParams,
      },
      { client },
    )
    return Marshaler.unwrapClients(response)
  }

  async updateById(
    id,
    patch,
    mask = Marshaler.fieldMaskFromPatch(
      patch,
      this._api.ClientRegistry.UpdateAllowedFieldMaskPaths,
    ),
  ) {
    const response = await this._api.ClientRegistry.Update(
      {
        routeParams: {
          'client.ids.client_id': id,
        },
      },
      {
        client: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )
    return Marshaler.unwrapClients(response)
  }

  async restoreById(id) {
    const response = await this._api.ClientRegistry.Restore({
      routeParams: {
        client_id: id,
      },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async deleteById(applicationId) {
    const response = await this._api.ClientRegistry.Delete({
      routeParams: { client_id: applicationId },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async purgeById(id) {
    const response = await this._api.ClientRegistry.Purge({
      routeParams: { client_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async getRightsById(id) {
    const result = await this._api.ClientAccess.ListRights({
      routeParams: { client_id: id },
    })

    return Marshaler.unwrapRights(result)
  }
}

export default Clients
