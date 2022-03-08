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

import autoBind from 'auto-bind'

import Marshaler from '../util/marshaler'

class PacketBrokerAgent {
  constructor(service) {
    this._api = service
    autoBind(this)
  }

  _emitDefault(idObject) {
    if (!('net_id' in idObject)) {
      return { ...idObject, net_id: 0 }
    }

    return idObject
  }

  _emitDefaults(iterable) {
    for (const probe of iterable) {
      if ('id' in probe) {
        probe.id = this._emitDefault(probe.id)
      } else {
        if ('forwarder_id' in probe) {
          probe.forwarder_id = this._emitDefault(probe.forwarder_id)
        }
        if ('home_network_id' in probe) {
          probe.home_network_id = this._emitDefault(probe.home_network_id)
        }
      }
    }

    return iterable
  }

  async getInfo() {
    const result = await this._api.GetInfo()

    return Marshaler.payloadSingleResponse(result)
  }

  async register(registration) {
    const result = await this._api.Register(undefined, registration)

    return Marshaler.payloadSingleResponse(result)
  }

  async deregister() {
    const result = await this._api.Deregister()

    return Marshaler.payloadSingleResponse(result)
  }

  async getHomeNetworkDefaultRoutingPolicy() {
    const result = await this._api.GetHomeNetworkDefaultRoutingPolicy()

    return Marshaler.payloadSingleResponse(result)
  }

  async setHomeNetworkDefaultRoutingPolicy(policy) {
    const result = await this._api.SetHomeNetworkDefaultRoutingPolicy(undefined, policy)

    return Marshaler.payloadSingleResponse(result)
  }

  async deleteHomeNetworkDefaultRoutingPolicy() {
    const result = await this._api.DeleteHomeNetworkDefaultRoutingPolicy()

    return Marshaler.payloadSingleResponse(result)
  }

  async getHomeNetworkRoutingPolicy(netId, tenantId) {
    const result = await this._api.GetHomeNetworkRoutingPolicy({
      routeParams: {
        net_id: netId,
        ...(tenantId ? { tenant_id: tenantId } : {}),
      },
    })

    return this._emitDefault(Marshaler.payloadSingleResponse(result))
  }

  async setHomeNetworkRoutingPolicy(netId, tenantId, policy) {
    const result = await this._api.SetHomeNetworkRoutingPolicy(
      {
        routeParams: {
          'home_network_id.net_id': netId,
          ...(tenantId ? { 'home_network_id.tenant_id': tenantId } : {}),
        },
      },
      policy,
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async deleteHomeNetworkRoutingPolicy(netId, tenantId) {
    const result = await this._api.DeleteHomeNetworkRoutingPolicy({
      routeParams: {
        net_id: netId,
        ...(tenantId ? { tenant_id: tenantId } : {}),
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async listHomeNetworkRoutingPolicies(params) {
    const result = await this._api.ListHomeNetworkRoutingPolicies(undefined, params)
    const response = Marshaler.unwrapPacketBrokerPolicies(result)
    response.policies = this._emitDefaults(response.policies)

    return response
  }

  async listNetworks(params) {
    const result = await this._api.ListNetworks(undefined, params)
    const response = Marshaler.unwrapPacketBrokerNetworks(result)
    response.networks = this._emitDefaults(response.networks)

    return response
  }

  async listForwarderRoutingPolicies(params) {
    const result = await this._api.ListForwarderRoutingPolicies(undefined, params)
    const response = Marshaler.unwrapPacketBrokerPolicies(result)
    response.policies = this._emitDefaults(response.policies)

    return response
  }

  async getHomeNetworkDefaultGatewayVisibility() {
    const result = await this._api.GetHomeNetworkDefaultGatewayVisibility()

    return Marshaler.payloadSingleResponse(result)
  }

  async setHomeNetworkDefaultGatewayVisibility(gatewayVisibility) {
    const result = await this._api.SetHomeNetworkDefaultGatewayVisibility(
      undefined,
      gatewayVisibility,
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async deleteHomeNetworkDefaultGatewayVisibility() {
    const result = await this._api.DeleteHomeNetworkDefaultGatewayVisibility()

    return Marshaler.payloadSingleResponse(result)
  }
}

export default PacketBrokerAgent
