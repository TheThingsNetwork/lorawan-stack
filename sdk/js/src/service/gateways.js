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
import stream from '../api/stream/stream-node'
import ApiKeys from './api-keys'

class Gateways {
  constructor (api, { defaultUserId, stackConfig, proxy = true }) {
    this._api = api
    this._defaultUserId = defaultUserId
    this._stackConfig = stackConfig
    this.ApiKeys = new ApiKeys(api.GatewayAccess, {
      parentRoutes: {
        list: 'gateway_ids.gateway_id',
        create: 'gateway_ids.gateway_id',
        update: 'gateway_ids.gateway_id',
      },
    })
  }

  // Retrieval

  async getAll (params, selector) {
    const response = await this._api.GatewayRegistry.List(undefined, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return Marshaler.unwrapGateways(response)
  }

  async getById (id, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.GatewayRegistry.Get({
      routeParams: { 'gateway_ids.gateway_id': id },
    }, fieldMask)

    return Marshaler.unwrapGateway(response)
  }

  // Update

  async updateById (id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    const response = await this._api.GatewayRegistry.Update({
      routeParams: { 'gateway.ids.gateway_id': id },
    },
    {
      gateway: patch,
      field_mask: Marshaler.fieldMask(mask),
    })

    return Marshaler.unwrapGateway(response)
  }

  // Create

  async create (userId = this._defaultUserId, gateway) {
    const response = await this._api.GatewayRegistry.Create({
      routeParams: { 'collaborator.user_ids.user_id': userId },
    },
    { gateway })

    return Marshaler.unwrapGateway(response)
  }

  // Delete

  async deleteById (id) {
    const response = await this._api.GatewayRegistry.Delete({
      routeParams: { gateway_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async getStatisticsById (id) {
    const response = await this._api.Gs.GetGatewayConnectionStats({
      routeParams: { gateway_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  // Events Stream

  async openStream (identifiers, tail, after) {
    const eventsUrl = `${this._stackConfig.as}/events`
    const payload = {
      identifiers: identifiers.map(id => ({
        gateway_ids: { gateway_id: id },
      })),
      tail,
      after,
    }

    return stream(payload, eventsUrl)
  }
}

export default Gateways
