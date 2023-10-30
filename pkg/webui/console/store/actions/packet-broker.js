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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationRequestActions,
  createPaginationBaseActionType,
} from '@ttn-lw/lib/store/actions/pagination'

export const SHARED_NAME = 'PACKET_BROKER_NETWORK'

export const GET_PACKET_BROKER_INFO_BASE = 'GET_PACKET_BROKER_INFO'
export const [
  {
    request: GET_PACKET_BROKER_INFO,
    success: GET_PACKET_BROKER_INFO_SUCCESS,
    failure: GET_PACKET_BROKER_INFO_FAILURE,
  },
  {
    request: getPacketBrokerInfo,
    success: getPacketBrokerInfoSuccess,
    failure: getPacketBrokerInfoFailure,
  },
] = createRequestActions(GET_PACKET_BROKER_INFO_BASE)

export const REGISTER_PACKET_BROKER_BASE = 'REGISTER_PACKET_BROKER'
export const [
  {
    request: REGISTER_PACKET_BROKER,
    success: REGISTER_PACKET_BROKER_SUCCESS,
    failure: REGISTER_PACKET_BROKER_FAILURE,
  },
  {
    request: registerPacketBroker,
    success: registerPacketBrokerSuccess,
    failure: registerPacketBrokerFailure,
  },
] = createRequestActions(REGISTER_PACKET_BROKER_BASE, registration => ({ registration }))

export const DEREGISTER_PACKET_BROKER_BASE = 'DEREGISTER_PACKET_BROKER'
export const [
  {
    request: DEREGISTER_PACKET_BROKER,
    success: DEREGISTER_PACKET_BROKER_SUCCESS,
    failure: DEREGISTER_PACKET_BROKER_FAILURE,
  },
  {
    request: deregisterPacketBroker,
    success: deregisterPacketBrokerSuccess,
    failure: deregisterPacketBrokerFailure,
  },
] = createRequestActions(DEREGISTER_PACKET_BROKER_BASE)

export const GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_BASE =
  'GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY'
export const [
  {
    request: GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
    success: GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
    failure: GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_FAILURE,
  },
  {
    request: getHomeNetworkDefaultRoutingPolicy,
    success: getHomeNetworkDefaultRoutingPolicySuccess,
    failure: getHomeNetworkDefaultRoutingPolicyFailure,
  },
] = createRequestActions(GET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_BASE)

export const SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_BASE =
  'SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY'
export const [
  {
    request: SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
    success: SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
    failure: SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_FAILURE,
  },
  {
    request: setHomeNetworkDefaultRoutingPolicy,
    success: setHomeNetworkDefaultRoutingPolicySuccess,
    failure: setHomeNetworkDefaultRoutingPolicyFailure,
  },
] = createRequestActions(SET_HOME_NETWORK_DEFAULT_ROUTING_POLICY_BASE, policy => ({ policy }))

export const DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY_BASE =
  'DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY'
export const [
  {
    request: DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY,
    success: DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY_SUCCESS,
    failure: DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY_FAILURE,
  },
  {
    request: deleteHomeNetworkDefaultRoutingPolicy,
    success: deleteHomeNetworkDefaultRoutingPolicySuccess,
    failure: deleteHomeNetworkDefaultRoutingPolicyFailure,
  },
] = createRequestActions(DELETE_HOME_NETWORK_DEFAULT_ROUTING_POLICY_BASE)

export const GET_PACKET_BROKER_NETWORK_BASE = 'GET_PACKET_BROKER_NETWORK'
export const [
  {
    request: GET_PACKET_BROKER_NETWORK,
    success: GET_PACKET_BROKER_NETWORK_SUCCESS,
    failure: GET_PACKET_BROKER_NETWORK_FAILURE,
  },
  {
    request: getPacketBrokerNetwork,
    success: getPacketBrokerNetworkSuccess,
    failure: getPacketBrokerNetworkFailure,
  },
] = createRequestActions(
  GET_PACKET_BROKER_NETWORK_BASE,
  id => ({ id }),
  (id, options) => ({ options }),
)

export const GET_PACKET_BROKER_NETWORKS_LIST_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  {
    request: GET_PACKET_BROKER_NETWORKS_LIST,
    success: GET_PACKET_BROKER_NETWORKS_LIST_SUCCESS,
    failure: GET_PACKET_BROKER_NETWORKS_LIST_FAILURE,
  },
  {
    request: getPacketBrokerNetworksList,
    success: getPacketBrokerNetworksListSuccess,
    failure: getPacketBrokerNetworksListFailure,
  },
] = createPaginationRequestActions(
  SHARED_NAME,
  ({ page, limit, query, order, withRoutingPolicy } = {}) => ({
    params: { page, limit, query, order, withRoutingPolicy },
  }),
)

export const GET_FORWARDER_ROUTING_POLICY_BASE = 'GET_FORWARDER_ROUTING_POLICY'
export const [
  {
    request: GET_FORWARDER_ROUTING_POLICY,
    success: GET_FORWARDER_ROUTING_POLICY_SUCCESS,
    failure: GET_FORWARDER_ROUTING_POLICY_FAILURE,
  },
  { request: getRoutingPolicy, success: getRoutingPolicySuccess, failure: getRoutingPolicyFailure },
] = createRequestActions(GET_FORWARDER_ROUTING_POLICY_BASE)

export const GET_FORWARDER_ROUTING_POLICIES_BASE = 'GET_FORWARDER_ROUTING_POLICIES'
export const [
  {
    request: GET_FORWARDER_ROUTING_POLICIES,
    success: GET_FORWARDER_ROUTING_POLICIES_SUCCESS,
    failure: GET_FORWARDER_ROUTING_POLICIES_FAILURE,
  },
  {
    request: getForwarderRoutingPolicies,
    success: getForwarderRoutingPoliciesSuccess,
    failure: getForwarderRoutingPoliciesFailure,
  },
] = createRequestActions(GET_FORWARDER_ROUTING_POLICIES_BASE)

export const GET_HOME_NETWORK_ROUTING_POLICY_BASE = 'GET_HOME_NETWORK_ROUTING_POLICY'
export const [
  {
    request: GET_HOME_NETWORK_ROUTING_POLICY,
    success: GET_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
    failure: GET_HOME_NETWORK_ROUTING_POLICY_FAILURE,
  },
  {
    request: getHomeNetworkRoutingPolicy,
    success: getHomeNetworkRoutingPolicySuccess,
    failure: getHomeNetworkRoutingPolicyFailure,
  },
] = createRequestActions(GET_HOME_NETWORK_ROUTING_POLICY_BASE)

export const SET_HOME_NETWORK_ROUTING_POLICY_BASE = 'SET_HOME_NETWORK_ROUTING_POLICY'
export const [
  {
    request: SET_HOME_NETWORK_ROUTING_POLICY,
    success: SET_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
    failure: SET_HOME_NETWORK_ROUTING_POLICY_FAILURE,
  },
  {
    request: setHomeNetworkRoutingPolicy,
    success: setHomeNetworkRoutingPolicySuccess,
    failure: setHomeNetworkRoutingPolicyFailure,
  },
] = createRequestActions(SET_HOME_NETWORK_ROUTING_POLICY_BASE, (id, policy) => ({ id, policy }))

export const DELETE_HOME_NETWORK_ROUTING_POLICY_BASE = 'DELETE_HOME_NETWORK_ROUTING_POLICY'
export const [
  {
    request: DELETE_HOME_NETWORK_ROUTING_POLICY,
    success: DELETE_HOME_NETWORK_ROUTING_POLICY_SUCCESS,
    failure: DELETE_HOME_NETWORK_ROUTING_POLICY_FAILURE,
  },
  {
    request: deleteHomeNetworkRoutingPolicy,
    success: deleteHomeNetworkRoutingPolicySuccess,
    failure: deleteHomeNetworkRoutingPolicyFailure,
  },
] = createRequestActions(DELETE_HOME_NETWORK_ROUTING_POLICY_BASE, (id, policy) => ({ id, policy }))

export const DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES_BASE =
  'DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES'
export const [
  {
    request: DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES,
    success: DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES_SUCCESS,
    failure: DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES_FAILURE,
  },
  {
    request: deleteAllHomeNetworkRoutingPolicies,
    success: deleteAllHomeNetworkRoutingPoliciesSuccess,
    failure: deleteAllHomeNetworkRoutingPoliciesFailure,
  },
] = createRequestActions(DELETE_ALL_HOME_NETWORK_ROUTING_POLICIES_BASE, ids => ({ ids }))

export const GET_HOME_NETWORK_ROUTING_POLICIES_BASE = 'GET_HOME_NETWORK_ROUTING_POLICIES'
export const [
  {
    request: GET_HOME_NETWORK_ROUTING_POLICIES,
    success: GET_HOME_NETWORK_ROUTING_POLICIES_SUCCESS,
    failure: GET_HOME_NETWORK_ROUTING_POLICIES_FAILURE,
  },
  {
    request: getHomeNetworkRoutingPolicies,
    success: getHomeNetworkRoutingPoliciesSuccess,
    failure: getHomeNetworkRoutingPoliciesFailure,
  },
] = createRequestActions(GET_HOME_NETWORK_ROUTING_POLICIES_BASE)

export const GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_BASE =
  'GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY'
export const [
  {
    request: GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY,
    success: GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_SUCCESS,
    failure: GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_FAILURE,
  },
  {
    request: getHomeNetworkDefaultGatewayVisibility,
    success: getHomeNetworkDefaultGatewayVisibilitySuccess,
    failure: getHomeNetworkDefaultGatewayVisibilityFailure,
  },
] = createRequestActions(GET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_BASE)

export const SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_BASE =
  'SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY'
export const [
  {
    request: SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY,
    success: SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_SUCCESS,
    failure: SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_FAILURE,
  },
  {
    request: setHomeNetworkDefaultGatewayVisibility,
    success: setHomeNetworkDefaultGatewayVisibilitySuccess,
    failure: setHomeNetworkDefaultGatewayVisibilityFailure,
  },
] = createRequestActions(SET_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_BASE, visibility => ({
  visibility,
}))

export const DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_BASE =
  'DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY'
export const [
  {
    request: DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY,
    success: DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_SUCCESS,
    failure: DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_FAILURE,
  },
  {
    request: deleteHomeNetworkDefaultGatewayVisibility,
    success: deleteHomeNetworkDefaultGatewayVisibilitySuccess,
    failure: deleteHomeNetworkDefaultGatewayVisibilityFailure,
  },
] = createRequestActions(DELETE_HOME_NETWORK_DEFAULT_GATEWAY_VISIBILITY_BASE)
