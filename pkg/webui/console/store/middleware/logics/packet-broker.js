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

import { defineMessages } from 'react-intl'

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import {
  extractPacketBrokerIdsFromCombinedId,
  combinePacketBrokerIds,
} from '@ttn-lw/lib/selectors/id'
import {
  isNotFoundError,
  isUnauthenticatedError,
  createFrontendError,
} from '@ttn-lw/lib/errors/utils'

import * as packetBroker from '@console/store/actions/packet-broker'

const m = defineMessages({
  unauthenticatedErrorTitle: 'Unable to authorize',
  unauthenticatedErrorMessage:
    'The Console is unable to authenticate requests to the Packet Broker agent. Please make sure the Packet Broker feature in The Things Stack is configured correctly. You can refer to the documentation above for further guidance.',
})

const unauthenticatedError = createFrontendError(
  m.unauthenticatedErrorTitle,
  m.unauthenticatedErrorMessage,
)

const fetchThroughPagination = async (endpoint, additionalArgs, process, shouldStop) => {
  let page = 1
  const limit = 100
  let totalCount = Infinity
  let stop = false
  let acc

  while ((page - 1) * limit < totalCount && !stop) {
    // eslint-disable-next-line no-await-in-loop
    const result = await endpoint({ page, limit, ...additionalArgs })

    acc = process(result, acc)
    stop = shouldStop ? shouldStop(result, acc) : false
    totalCount = result.totalCount
    page += 1
  }

  return acc
}

const getDefaultRoutingPolicy = async () => {
  try {
    return await tts.PacketBrokerAgent.getHomeNetworkDefaultRoutingPolicy()
  } catch (error) {
    // Not found error means that no policy is set.
    if (isNotFoundError(error)) {
      return {}
    }

    throw error
  }
}

const getPacketBrokerInfoLogic = createRequestLogic({
  type: packetBroker.GET_PACKET_BROKER_INFO,
  process: async () => {
    try {
      return await tts.PacketBrokerAgent.getInfo()
    } catch (error) {
      if (isUnauthenticatedError(error)) {
        // Faulty configurations can lead to 401, which in turn trigger
        // the page refresh of the request logic. These errors need to be
        // intercepted to prevent that.

        return unauthenticatedError
      }

      throw error
    }
  },
})

const registerPacketBrokerLogic = createRequestLogic({
  type: packetBroker.REGISTER_PACKET_BROKER,
  process: async ({ action }) => {
    const { registration } = action.payload

    return await tts.PacketBrokerAgent.register(registration)
  },
})

const deregisterPacketBrokerLogic = createRequestLogic({
  type: packetBroker.DEREGISTER_PACKET_BROKER,
  process: tts.PacketBrokerAgent.deregister,
})

const getHomeNetworkDefaultRoutingPolicyLogic = createRequestLogic({
  type: packetBroker.GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
  process: getDefaultRoutingPolicy,
})

const setHomeNetworkDefaultRoutingPolicyLogic = createRequestLogic({
  type: packetBroker.SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
  process: async ({ action }) => {
    const { policy } = action.payload

    await tts.PacketBrokerAgent.setHomeNetworkDefaultRoutingPolicy(policy)

    return policy
  },
})

const deleteHomeNetworkDefaultRoutingPolicyLogic = createRequestLogic({
  type: packetBroker.DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
  process: async () => {
    try {
      await tts.PacketBrokerAgent.deleteHomeNetworkDefaultRoutingPolicy()
    } catch (error) {
      // We can ignore not found errors, meaning that the
      // policy was already deleted or never existed.
      if (!isNotFoundError(error)) {
        throw error
      }
    }
  },
})

const getPacketBrokerNetworkLogic = createRequestLogic({
  type: packetBroker.GET_PACKET_BROKER_NETWORK,
  process: async ({ action }, dispatch) => {
    const {
      payload: { id },
      meta: { options: { fetchPolicies } = {} },
    } = action
    const ids = extractPacketBrokerIdsFromCombinedId(id)

    const network = await fetchThroughPagination(
      tts.PacketBrokerAgent.listNetworks,
      ids.tenant_id ? { tenant_id_contains: ids.tenant_id } : undefined,
      result => result.networks.find(n => combinePacketBrokerIds(n.id) === id),
      (result, acc) => Boolean(acc),
    )

    if (!network) {
      throw { statusCode: 404 }
    }

    if (network && fetchPolicies) {
      const fetchHomeNetworkRoutingPolicy = async () => {
        try {
          return await tts.PacketBrokerAgent.getHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id)
        } catch (error) {
          if (isNotFoundError(error)) {
            return { home_network_id: ids }
          }
          throw error
        }
      }
      const [forwarder, homeNetwork] = await Promise.all([
        fetchThroughPagination(
          tts.PacketBrokerAgent.listForwarderRoutingPolicies,
          undefined,
          (result, acc = []) => [...acc, ...result.policies],
        ),
        fetchHomeNetworkRoutingPolicy(),
      ])
      dispatch(packetBroker.getForwarderRoutingPoliciesSuccess(forwarder))
      dispatch(packetBroker.getHomeNetworkRoutingPolicySuccess(homeNetwork))
    }

    return network
  },
})

const getPacketBrokerNetworksLogic = createRequestLogic({
  type: packetBroker.GET_PACKET_BROKER_NETWORKS_LIST,
  latest: true,
  process: async ({ action }, dispatch) => {
    const {
      payload: {
        params: { page, limit, query, withRoutingPolicy },
      },
      meta: {
        options: { fetchPolicies = {} },
      },
    } = action

    const data = await tts.PacketBrokerAgent.listNetworks({
      page,
      limit,
      name_contains: query,
      tenant_id_contains: query,
      with_routing_policy: withRoutingPolicy ? 'true' : 'false',
    })

    if (fetchPolicies) {
      const [defaultPolicy, forwarder, homeNetwork] = await Promise.all([
        getDefaultRoutingPolicy(),
        fetchThroughPagination(
          tts.PacketBrokerAgent.listForwarderRoutingPolicies,
          undefined,
          (result, acc = []) => [...acc, ...result.policies],
        ),
        fetchThroughPagination(
          tts.PacketBrokerAgent.listHomeNetworkRoutingPolicies,
          undefined,
          (result, acc = []) => [...acc, ...result.policies],
        ),
      ])
      dispatch(packetBroker.getHomeNetworkDefaultRoutingPolicySuccess(defaultPolicy))
      dispatch(packetBroker.getForwarderRoutingPoliciesSuccess(forwarder))
      dispatch(packetBroker.getHomeNetworkRoutingPoliciesSuccess(homeNetwork))
    }

    return { entities: data.networks, totalCount: data.totalCount }
  },
})

const getPacketBrokerForwarderPoliciesLogic = createRequestLogic({
  type: packetBroker.GET_FORWARDER_ROUTING_POLICIES,
  process: async () => {
    const data = fetchThroughPagination(
      tts.PacketBrokerAgent.listForwarderRoutingPolicies,
      undefined,
      (result, acc = []) => [...acc, ...result.policies],
    )

    return data
  },
})

const getPacketBrokerHomeNetworkPoliciesLogic = createRequestLogic({
  type: packetBroker.GET_HOME_NETWORK_ROUTING_POLICIES,
  process: async () => {
    const data = fetchThroughPagination(
      tts.PacketBrokerAgent.listHomeNetworkRoutingPolicies,
      undefined,
      (result, acc = []) => [...acc, ...result.policies],
    )

    return data
  },
})

const setPacketBrokerHomeNetworkPolicyLogic = createRequestLogic({
  type: packetBroker.SET_HOME_NETWORK_ROUTING_POLICY,
  process: async ({ action }) => {
    const {
      payload: { id, policy },
    } = action
    const ids = extractPacketBrokerIdsFromCombinedId(id)
    await tts.PacketBrokerAgent.setHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id, policy)

    const newPolicy = { home_network_id: { net_id: ids.net_id }, ...policy }
    if ('tenant_id' in ids) {
      newPolicy.home_network_id.tenant_id = ids.tenant_id
    }

    return newPolicy
  },
})

const deletePacketBrokerHomeNetworkPolicyLogic = createRequestLogic({
  type: packetBroker.DELETE_HOME_NETWORK_ROUTING_POLICY,
  process: async ({ action }) => {
    const {
      payload: { id },
    } = action
    const ids = extractPacketBrokerIdsFromCombinedId(id)

    try {
      await tts.PacketBrokerAgent.deleteHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id)
    } catch (error) {
      // We can ignore not found errors, meaning that the
      // policy was already deleted or never existed.
      if (!isNotFoundError(error)) {
        throw error
      }
    }

    return ids
  },
})

const deleteAllPacketBrokerHomeNetworkPoliciesLogic = createRequestLogic({
  type: packetBroker.DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES,
  process: async ({ action }) => {
    const {
      payload: { ids },
    } = action

    try {
      await Promise.all(
        ids.map(async id => {
          const ids = extractPacketBrokerIdsFromCombinedId(id)
          if (typeof ids === 'number') {
            return tts.PacketBrokerAgent.deleteHomeNetworkRoutingPolicy(ids)
          }

          return tts.PacketBrokerAgent.deleteHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id)
        }),
      )

      return ids
    } catch (error) {
      if (!isNotFoundError(error)) {
        throw error
      }
    }
  },
})

const getDefaultGatewayVisibilityLogic = createRequestLogic({
  type: packetBroker.GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY,
  process: async () => {
    try {
      const result = await tts.PacketBrokerAgent.getHomeNetworkDefaultGatewayVisibility()

      return result
    } catch (error) {
      // Not found error means that no gateway visibility is set.
      if (isNotFoundError(error)) {
        return {}
      }

      throw error
    }
  },
})

const setDefaultGatewayVisibilityLogic = createRequestLogic({
  type: packetBroker.SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY,
  process: async ({ action }) => {
    const { payload } = action

    await tts.PacketBrokerAgent.setHomeNetworkDefaultGatewayVisibility(payload)

    return payload
  },
})

const deleteDefaultGatewayVisibilityLogic = createRequestLogic({
  type: packetBroker.DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY,
  process: async () => {
    try {
      await tts.PacketBrokerAgent.deleteHomeNetworkDefaultGatewayVisibility()
    } catch (error) {
      // We can ignore not found errors, meaning that the
      // gateway visibility was already deleted or never existed.
      if (!isNotFoundError(error)) {
        throw error
      }
    }
  },
})

export default [
  getPacketBrokerInfoLogic,
  registerPacketBrokerLogic,
  deregisterPacketBrokerLogic,
  getHomeNetworkDefaultRoutingPolicyLogic,
  setHomeNetworkDefaultRoutingPolicyLogic,
  deleteHomeNetworkDefaultRoutingPolicyLogic,
  getPacketBrokerNetworkLogic,
  getPacketBrokerNetworksLogic,
  getPacketBrokerForwarderPoliciesLogic,
  getPacketBrokerHomeNetworkPoliciesLogic,
  setPacketBrokerHomeNetworkPolicyLogic,
  deletePacketBrokerHomeNetworkPolicyLogic,
  deleteAllPacketBrokerHomeNetworkPoliciesLogic,
  getDefaultGatewayVisibilityLogic,
  setDefaultGatewayVisibilityLogic,
  deleteDefaultGatewayVisibilityLogic,
]
