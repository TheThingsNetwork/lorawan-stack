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
import { getComponentsWithDistinctBaseUrls } from '../util/stack-components'
import combineStreams from '../util/combine-streams'
import ApiKeys from './api-keys'
import Collaborators from './collaborators'

class Gateways {
  constructor(api, { defaultUserId, stackConfig, proxy = true }) {
    this._api = api
    this._defaultUserId = defaultUserId
    this._stackConfig = stackConfig
    this.ApiKeys = new ApiKeys(api.GatewayAccess, {
      parentRoutes: {
        get: 'gateway_ids.gateway_id',
        list: 'gateway_ids.gateway_id',
        create: 'gateway_ids.gateway_id',
        update: 'gateway_ids.gateway_id',
      },
    })
    this.Collaborators = new Collaborators(api.GatewayAccess, {
      parentRoutes: {
        get: 'gateway_ids.gateway_id',
        list: 'gateway_ids.gateway_id',
        set: 'gateway_ids.gateway_id',
      },
    })
  }

  // Retrieval

  async getAll(params, selector) {
    const response = await this._api.GatewayRegistry.List(undefined, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return Marshaler.unwrapGateways(response)
  }

  async getById(id, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.GatewayRegistry.Get(
      {
        routeParams: { 'gateway_ids.gateway_id': id },
      },
      fieldMask,
    )

    return Marshaler.unwrapGateway(response)
  }

  // Update

  async updateById(id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    const response = await this._api.GatewayRegistry.Update(
      {
        routeParams: { 'gateway.ids.gateway_id': id },
      },
      {
        gateway: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )

    return Marshaler.unwrapGateway(response)
  }

  // Create

  async create(ownerId = this._defaultUserId, gateway, isUserOwner = true) {
    const routeParams = isUserOwner
      ? { 'collaborator.user_ids.user_id': ownerId }
      : { 'collaborator.organization_ids.organization_id': ownerId }
    const response = await this._api.GatewayRegistry.Create(
      {
        routeParams,
      },
      { gateway },
    )

    return Marshaler.unwrapGateway(response)
  }

  // Delete

  async deleteById(id) {
    const response = await this._api.GatewayRegistry.Delete({
      routeParams: { gateway_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async getStatisticsById(id) {
    const response = await this._api.Gs.GetGatewayConnectionStats({
      routeParams: { gateway_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async getRightsById(gatewayId) {
    const result = await this._api.GatewayAccess.ListRights({
      routeParams: { gateway_id: gatewayId },
    })

    return Marshaler.unwrapRights(result)
  }

  // Events Stream

  async openStream(identifiers, tail, after) {
    const payload = {
      identifiers: identifiers.map(id => ({
        gateway_ids: { gateway_id: id },
      })),
      tail,
      after,
    }

    // Event streams can come from multiple stack components. It is necessary to
    // check for stack components on different hosts and open distinct stream
    // connections for any distinct host if need be.
    const distinctComponents = getComponentsWithDistinctBaseUrls(this._stackConfig, ['is', 'gs'])

    const streams = distinctComponents.map(component =>
      this._api.Events.Stream({ component }, payload),
    )

    // Combine all stream sources to one subscription generator.
    return combineStreams(streams)
  }
}

export default Gateways
