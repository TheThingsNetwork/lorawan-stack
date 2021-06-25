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

import api from '@console/api'

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
  const limit = 1000
  let totalCount = Infinity
  let stop = false
  let acc

  while (page * limit <= totalCount && !stop) {
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
    return await api.packetBroker.getHomeNetworkDefaultRoutingPolicy()
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
      return await api.packetBroker.getInfo()
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
  process: api.packetBroker.register,
})

const deregisterPacketBrokerLogic = createRequestLogic({
  type: packetBroker.DEREGISTER_PACKET_BROKER,
  process: api.packetBroker.deregister,
})

const getHomeNetworkDefaultRoutingPolicyLogic = createRequestLogic({
  type: packetBroker.GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
  process: getDefaultRoutingPolicy,
})

const setHomeNetworkDefaultRoutingPolicyLogic = createRequestLogic({
  type: packetBroker.SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
  process: async ({ action }) => {
    const { policy } = action.payload

    await api.packetBroker.setHomeNetworkDefaultRoutingPolicy(policy)

    return policy
  },
})

const deleteHomeNetworkDefaultRoutingPolicyLogic = createRequestLogic({
  type: packetBroker.DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
  process: async () => {
    try {
      await api.packetBroker.deleteHomeNetworkDefaultRoutingPolicy()
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
      api.packetBroker.listNetworks,
      ids.tenant_id ? { tenant_id_contains: ids.tenant_id } : undefined,
      (result, acc) => result.networks.find(n => combinePacketBrokerIds(n.id) === id),
      (result, acc) => Boolean(acc),
    )

    if (network && fetchPolicies) {
      const fetchHomeNetworkRoutingPolicy = async () => {
        try {
          return await api.packetBroker.getHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id)
        } catch (error) {
          if (isNotFoundError(error)) {
            return { home_network_id: ids }
          }
          throw error
        }
      }
      const [forwarder, homeNetwork] = await Promise.all([
        fetchThroughPagination(
          api.packetBroker.listForwarderRoutingPolicies,
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

    const data = await api.packetBroker.listNetworks({
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
          api.packetBroker.listForwarderRoutingPolicies,
          undefined,
          (result, acc = []) => [...acc, ...result.policies],
        ),
        fetchThroughPagination(
          api.packetBroker.listHomeNetworkRoutingPolicies,
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
      api.packetBroker.listForwarderRoutingPolicies,
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
      api.packetBroker.listHomeNetworkRoutingPolicies,
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
    await api.packetBroker.setHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id, policy)

    return policy
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
      await api.packetBroker.deleteHomeNetworkRoutingPolicy(ids.net_id, ids.tenant_id)
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
]
