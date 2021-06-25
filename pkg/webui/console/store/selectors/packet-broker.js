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

import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import { combinePacketBrokerIds } from '@ttn-lw/lib/selectors/id'

import {
  GET_PACKET_BROKER_INFO_BASE,
  GET_PACKET_BROKER_NETWORKS_LIST_BASE,
} from '@console/store/actions/packet-broker'

const ENTITY = 'packetBrokerNetworks'

const selectPacketBrokerStore = state => state.packetBroker

// General.
export const selectInfo = state => selectPacketBrokerStore(state).info
export const selectRegistration = state => selectInfo(state).registration || {}
export const selectPacketBrokerOwnCombinedId = state =>
  combinePacketBrokerIds(selectRegistration(state).id)
export const selectInfoFetching = createFetchingSelector(GET_PACKET_BROKER_INFO_BASE)
export const selectInfoError = createErrorSelector(GET_PACKET_BROKER_INFO_BASE)

export const selectRegistered = state => selectPacketBrokerStore(state).registered
export const selectEnabled = state => selectPacketBrokerStore(state).enabled

export const selectHomeNetworkDefaultRoutingPolicy = state =>
  selectPacketBrokerStore(state).defaultHomeNetworkRoutingPolicy

// Network.
export const selectPacketBrokerNetworkEntitiesStore = state =>
  selectPacketBrokerStore(state).networks.entities
export const selectPacketBrokerNetworkById = (state, combinedId) =>
  selectPacketBrokerNetworkEntitiesStore(state)[combinedId]

// Networks.
const selectPBNetworksIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectPBNetworksTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectPBNetworksFetching = createFetchingSelector(GET_PACKET_BROKER_NETWORKS_LIST_BASE)
const selectPBNetworksError = createErrorSelector(GET_PACKET_BROKER_NETWORKS_LIST_BASE)

export const selectPacketBrokerNetworks = state =>
  selectPBNetworksIds(state).map(netId => selectPacketBrokerNetworkById(state, netId))
export const selectPacketBrokerNetworksTotalCount = state => selectPBNetworksTotalCount(state)
export const selectPacketBrokerNetworksFetching = state => selectPBNetworksFetching(state)
export const selectPacketBrokerNetworksError = state => selectPBNetworksError(state)

// Policies.
export const selectPacketBrokerPoliciesStore = state => selectPacketBrokerStore(state).policies
export const selectPacketBrokerForwarderPoliciesStore = state =>
  selectPacketBrokerPoliciesStore(state).forwarders
export const selectPacketBrokerForwarderPolicyById = (state, combinedId) =>
  selectPacketBrokerForwarderPoliciesStore(state)[combinedId]
export const selectPacketBrokerHomeNetworkPoliciesStore = state =>
  selectPacketBrokerPoliciesStore(state).homeNetworks
export const selectPacketBrokerHomeNetworkPolicyById = (state, combinedId) =>
  selectPacketBrokerHomeNetworkPoliciesStore(state)[combinedId]
