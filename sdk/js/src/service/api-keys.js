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

class ApiKeys {
  constructor (registry, { parentRoutes }) {
    this._api = registry
    this._parentRoutes = parentRoutes
  }

  async getById (entityId, id) {
    const entityIdRoute = this._parentRoutes.get
    const result = await this._api.GetAPIKey({
      routeParams: { [entityIdRoute]: entityId, key_id: id },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getAll (entityId, params) {
    const entityIdRoute = this._parentRoutes.list
    const result = await this._api.ListAPIKeys({
      routeParams: { [entityIdRoute]: entityId },
    }, params)

    return Marshaler.payloadListResponse('api_keys', result)
  }

  async create (entityId, key) {
    const entityIdRoute = this._parentRoutes.create
    const result = await this._api.CreateAPIKey({
      routeParams: { [entityIdRoute]: entityId },
    },
    {
      ...key,
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async deleteById (entityId, id) {
    return this.updateById(entityId, id, {
      rights: [],
    })
  }

  async updateById (entityId, id, patch) {
    const entityIdRoute = this._parentRoutes.update
    const result = await this._api.UpdateAPIKey({
      routeParams: {
        [entityIdRoute]: entityId,
        'api_key.id': id,
      },
    },
    {
      api_key: { ...patch },
    })

    return Marshaler.payloadSingleResponse(result)
  }
}

export default ApiKeys
