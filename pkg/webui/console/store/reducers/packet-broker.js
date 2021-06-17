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

import { handleActions, combineActions } from 'redux-actions'

import {
  getPacketBrokerNetworkId,
  getPacketBrokerHomeNewtorkId,
  getPacketBrokerForwarderId,
} from '@ttn-lw/lib/selectors/id'

import {
  GET_PACKET_BROKER_INFO_SUCCESS,
  REGISTER_PACKET_BROKER_SUCCESS,
  DEREGISTER_PACKET_BROKER_SUCCESS,
  GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
  SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
  DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
  GET_PACKET_BROKER_NETWORKS_LIST_SUCCESS,
  GET_FORWARDER_ROUTING_POLICIES_SUCCESS,
  GET_HOME_NETWORK_ROUTING_POLICIES_SUCCESS,
  GET_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
  GET_PACKET_BROKER_NETWORK_SUCCESS,
  SET_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
  DELETE_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
} from '@console/store/actions/packet-broker'

const defaultState = {
  info: {},
  registered: false,
  enabled: false,
  defaultHomeNetworkRoutingPolicy: {},
  networks: {
    entities: {},
  },
  policies: {
    forwarders: {},
    homeNetworks: {},
  },
}

const addPolicy = (state, policy, isForwarder) => {
  const id = isForwarder ? getPacketBrokerForwarderId(policy) : getPacketBrokerHomeNewtorkId(policy)
  return {
    ...state,
    [id]: {
      ...(state[id] || {}),
      ...policy,
    },
  }
}

export default handleActions(
  {
    [GET_PACKET_BROKER_INFO_SUCCESS]: (state, { payload }) => ({
      ...state,
      info: payload,
      registered: Boolean(payload.registration),
      enabled: true,
    }),
    [REGISTER_PACKET_BROKER_SUCCESS]: state => ({
      ...state,
      registered: true,
    }),
    [DEREGISTER_PACKET_BROKER_SUCCESS]: state => ({
      ...state,
      registered: false,
    }),
    [GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS]: (state, { payload }) => ({
      ...state,
      defaultHomeNetworkRoutingPolicy: payload,
    }),
    [SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS]: (state, { payload }) => ({
      ...state,
      defaultHomeNetworkRoutingPolicy: payload,
    }),
    [combineActions(
      DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
      DEREGISTER_PACKET_BROKER_SUCCESS,
    )]: state => ({
      ...state,
      defaultHomeNetworkRoutingPolicy: defaultState.defaultHomeNetworkRoutingPolicy,
    }),
    [GET_PACKET_BROKER_NETWORK_SUCCESS]: (state, { payload }) => {
      const id = getPacketBrokerNetworkId(payload)
      return {
        ...state,
        networks: {
          ...state.networks,
          entities: {
            ...state.networks.entities,
            [id]: {
              ...(state.networks[id] || {}),
              ...payload,
            },
          },
        },
      }
    },
    [GET_PACKET_BROKER_NETWORKS_LIST_SUCCESS]: (state, { payload }) => {
      const entities = payload.entities.reduce(
        (acc, nwk) => {
          const id = getPacketBrokerNetworkId(nwk)
          acc[id] = {
            ...(acc[id] || {}),
            ...nwk,
          }
          return acc
        },
        { ...state.networks.entities },
      )
      return {
        ...state,
        networks: {
          ...state.networks,
          entities,
        },
      }
    },
    [GET_FORWARDER_ROUTING_POLICIES_SUCCESS]: (state, { payload: policies }) => ({
      ...state,
      policies: {
        ...state.policies,
        forwarders: policies.reduce((acc, policy) => addPolicy(acc, policy, true), {}),
      },
    }),
    [combineActions(
      GET_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
      SET_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
    )]: (state, { payload: policy }) => ({
      ...state,
      policies: {
        ...state.policies,
        homeNetworks: addPolicy(state.policies.homeNetworks, policy, false),
      },
    }),
    [DELETE_HOME_NETWORK_ROUTING_POLICY_SUCCESS]: (state, { payload: id }) => ({
      ...state,
      policies: {
        ...state.policies,
        homeNetworks: addPolicy(state.policies.homeNetworks, { home_network_id: id }, false),
      },
    }),
    [GET_HOME_NETWORK_ROUTING_POLICIES_SUCCESS]: (
      state,
      { payload: policies },
      getPacketBrokerHomeNewtorkId,
    ) => ({
      ...state,
      policies: {
        ...state.policies,
        homeNetworks: policies.reduce((acc, policy) => addPolicy(acc, policy, false), {}),
      },
    }),
  },
  defaultState,
)
