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
  GET_GTWS_LIST_BASE,
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
import {
  createApiKeysSelector,
  createApiKeysStoreSelector,
} from './api-keys'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'
import { createApiKeySelector } from './api-key'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'gateways'
const ENTITY_SINGLE = 'gateway'


// Gateway Entity
export const selectGatewayStore = state => state.gateways
export const selectGatewayEntitiesStore = state => selectGatewayStore(state).entities
export const selectGatewayStatisticsStore = state => selectGatewayStore(state).statistics
export const selectGatewayById = (state, id) => selectGatewayEntitiesStore(state)[id]
export const selectSelectedGatewayId = state => selectGatewayStore(state).selectedGateway
export const selectSelectedGateway = state => selectGatewayById(state, selectSelectedGatewayId(state))
export const selectGatewayFetching = createFetchingSelector(GET_GTW_BASE)
export const selectGatewayError = createErrorSelector(GET_GTW_BASE)

// Gateways
const selectGtwsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectGtwsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectGtwsFetching = createFetchingSelector(GET_GTWS_LIST_BASE)
const selectGtwsError = createErrorSelector(GET_GTWS_LIST_BASE)

export const selectGateways = state => selectGtwsIds(state).map(id => selectGatewayById(state, id))
export const selectGatewaysTotalCount = state => selectGtwsTotalCount(state)
export const selectGatewaysFetching = state => selectGtwsFetching(state)
export const selectGatewaysError = state => selectGtwsError(state)

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
export const selectGatewayStatisticsById = (state, id) => selectGatewayStatisticsStore(state)[id] || {}

export const selectGatewayStatisticsError = createErrorSelector(UPDATE_GTW_STATS_BASE)
export const selectGatewayStatisticsIsFetching = createFetchingSelector(UPDATE_GTW_STATS_BASE)
export const selectGatewayStatistics = function (state) {
  const statistics = selectGatewayStatisticsById(state, selectSelectedGatewayId(state))

  return statistics.stats
}
export const selectGatewayStatisticsIsAvailable = function (state) {
  const statistics = selectGatewayStatisticsById(state, selectSelectedGatewayId(state))

  return statistics.available || false
}
