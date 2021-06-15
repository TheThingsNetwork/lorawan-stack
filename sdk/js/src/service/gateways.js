// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import combineStreams from '../util/combine-streams'
import { STACK_COMPONENTS_MAP } from '../util/constants'

import ApiKeys from './api-keys'
import Collaborators from './collaborators'

class Gateways {
  constructor(api, { defaultUserId, stackConfig }) {
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

  _emitDefaults(paths, gateway) {
    // Handle zero coordinates that are swallowed by the grpc-gateway for
    // gateway antennas.
    if (paths.includes('antennas') && Boolean(gateway.antennas)) {
      const { antennas } = gateway

      for (const antenna of antennas) {
        if (!('altitude' in antenna.location)) {
          antenna.location.altitude = 0
        }
        if (!('longitude' in antenna.location)) {
          antenna.location.longitude = 0
        }
        if (!('latitude' in antenna.location)) {
          antenna.location.latitude = 0
        }
      }
    }

    // Handle missing boolean values.
    if (paths.includes('location_public') && !Boolean(gateway.location_public)) {
      gateway.location_public = false
    }
    if (paths.includes('auto_update') && !Boolean(gateway.auto_update)) {
      gateway.auto_update = false
    }
    if (
      paths.includes('require_authenticated_connection') &&
      !Boolean(gateway.require_authenticated_connection)
    ) {
      gateway.require_authenticated_connection = false
    }
    if (
      paths.includes('update_location_from_status') &&
      !Boolean(gateway.update_location_from_status)
    ) {
      gateway.update_location_from_status = false
    }

    return gateway
  }

  // Retrieval.

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

    return this._emitDefaults(fieldMask.field_mask.paths, Marshaler.unwrapGateway(response))
  }

  async search(params, selector) {
    const response = await this._api.EntityRegistrySearch.SearchGateways(undefined, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return Marshaler.unwrapGateways(response)
  }

  // Update.

  async updateById(
    id,
    patch,
    mask = Marshaler.fieldMaskFromPatch(
      patch,
      this._api.GatewayRegistry.UpdateAllowedFieldMaskPaths,
    ),
  ) {
    const response = await this._api.GatewayRegistry.Update(
      {
        routeParams: { 'gateway.ids.gateway_id': id },
      },
      {
        gateway: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )

    return this._emitDefaults(mask, Marshaler.unwrapGateway(response))
  }

  // Creation.

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

  // Deletion.

  async deleteById(id) {
    const response = await this._api.GatewayRegistry.Delete({
      routeParams: { gateway_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async purgeById(id) {
    const response = await this._api.GatewayRegistry.Purge({
      routeParams: { gateway_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  // Miscellaneous.

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
    const distinctComponents = this._stackConfig.getComponentsWithDistinctBaseUrls([
      STACK_COMPONENTS_MAP.is,
      STACK_COMPONENTS_MAP.gs,
    ])

    const streams = distinctComponents.map(component =>
      this._api.Events.Stream({ component }, payload),
    )

    // Combine all stream sources to one subscription generator.
    return combineStreams(streams)
  }

  // Gateway Configuration Server.

  async getGlobalConf(gatewayId) {
    // Endpoint hardcoded because it is not part of the gRPC API.
    // Refactor implementation once the following issue is resolved:
    // https://github.com/TheThingsNetwork/lorawan-stack/issues/3280
    const endpoint = `/gcs/gateways/${gatewayId}/semtechudp/global_conf.json`

    const response = await this._api._connector.handleRequest('get', endpoint, 'gcs', false)

    return Marshaler.payloadSingleResponse(response.data)
  }
}

export default Gateways
