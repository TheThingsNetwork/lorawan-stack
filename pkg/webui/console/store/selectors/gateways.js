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

import {
  GET_GTW_BASE,
  GET_GTW_API_KEY_BASE,
  GET_GTW_API_KEYS_LIST_BASE,
  UPDATE_GTW_STATS_BASE,
  GET_GTWS_RIGHTS_LIST_BASE,
} from '../actions/gateways'

import { getGatewayId } from '../../../lib/selectors/id'
import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
} from './events'
import {
  createRightsSelector,
  createUniversalRightsSelector,
} from './rights'
import { createApiKeysSelector, createApiKeysStoreSelector } from './api-keys'
import { createApiKeySelector } from './api-key'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'gateways'
const ENTITY_SINGLE = 'gateway'

const selectGatewayStore = state => state.gateway

// Gateway Entity
export const selectSelectedGateway = state => selectGatewayStore(state).gateway
export const selectSelectedGatewayId = state => getGatewayId(selectSelectedGateway(state))
export const selectGatewayFetching = createFetchingSelector(GET_GTW_BASE)
export const selectGatewayError = createErrorSelector(GET_GTW_BASE)

// Events
export const selectGatewayEvents = createEventsSelector(ENTITY)
export const selectGatewayEventsError = createEventsErrorSelector(ENTITY)
export const selectGatewayEventsStatus = createEventsStatusSelector(ENTITY)

// Api Keys
export const selectGatewayApiKeysStore = createApiKeysStoreSelector(ENTITY)
export const selectGatewayApiKeys = createApiKeysSelector(ENTITY)
export const selectGatewayKeysError = createErrorSelector(GET_GTW_API_KEYS_LIST_BASE)
export const selectGatewayKeysFetching = createFetchingSelector(GET_GTW_API_KEYS_LIST_BASE)
export const selectGatewayApiKey = createApiKeySelector(ENTITY_SINGLE)
export const selectGatewayApiKeyFetching = createFetchingSelector(GET_GTW_API_KEY_BASE)
export const selectGatewayApiKeyError = createErrorSelector(GET_GTW_API_KEY_BASE)

// Rights
export const selectGatewayRights = createRightsSelector(ENTITY)
export const selectGatewayUniversalRights = createUniversalRightsSelector(ENTITY)
export const selectGatewayRightsError = createErrorSelector(ENTITY)
export const selectGatewayRightsFetching = createFetchingSelector(GET_GTWS_RIGHTS_LIST_BASE)

// Statistics
export const selectGatewayStatisticsError = createErrorSelector(UPDATE_GTW_STATS_BASE)
export const selectGatewayStatisticsIsFetching = createFetchingSelector(UPDATE_GTW_STATS_BASE)
const selectGatewayStatisticStore = function (state) {
  const store = selectGatewayStore(state)

  return store.statistics
}
export const selectGatewayStatistics = function (state) {
  const store = selectGatewayStatisticStore(state)

  return store.stats
}
export const selectGatewayStatisticsIsAvailable = function (state) {
  const store = selectGatewayStatisticStore(state)

  return store.available
}
