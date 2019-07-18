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

class Collaborators {
  constructor (registry, { parentRoutes }) {
    this._api = registry
    this._parentRoutes = parentRoutes
  }

  async _getById (entityId, collaboratorId, isUser) {
    const entityIdRoute = this._parentRoutes.get
    const collaboratorIdRoute = isUser
      ? 'collaborator.user_ids.user_id'
      : 'collaborator.organization_ids.organization_id'

    const result = await this._api.GetCollaborator({
      routeParams: {
        [entityIdRoute]: entityId,
        [collaboratorIdRoute]: collaboratorId,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getByUserId (entityId, userId) {
    return this._getById(entityId, userId, true)
  }

  async getByOrganizationId (entityId, organizationId) {
    return this._getById(entityId, organizationId, false)
  }

  async getAll (entityId) {
    const entityIdRoute = this._parentRoutes.list
    const result = await this._api.ListCollaborators({
      routeParams: { [entityIdRoute]: entityId },
    })

    return Marshaler.payloadListResponse('collaborators', result)
  }

  async add (entityId, data) {
    const entityIdRoute = this._parentRoutes.set
    const result = await this._api.SetCollaborator({
      routeParams: { [entityIdRoute]: entityId },
    },
    {
      collaborator: data,
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async update (entityId, data) {
    return await this.add(entityId, data)
  }

  async remove (entityId, data) {
    return await this.add(entityId, { ...data, rights: []})
  }
}

export default Collaborators
