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

import { createSelector } from 'reselect'

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'

import {
  UPDATE_GTW_STATS_BASE,
  GET_GTWS_RIGHTS_LIST_BASE,
  START_GTW_STATS_BASE,
} from '@console/store/actions/gateways'

import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
  createEventsInterruptedSelector,
  createEventsPausedSelector,
  createEventsTruncatedSelector,
  createLatestEventSelector,
  createEventsFilterSelector,
} from './events'
import { createRightsSelector, createPseudoRightsSelector } from './rights'

const ENTITY = 'gateways'

// Gateway.
export const selectGatewayStore = state => state.gateways
export const selectGatewayEntitiesStore = state => selectGatewayStore(state).entities
export const selectGatewayStatisticsStore = state => selectGatewayStore(state).statistics
export const selectGatewayById = (state, id) => selectGatewayEntitiesStore(state)[id]
export const selectSelectedGatewayId = state => selectGatewayStore(state).selectedGateway
export const selectSelectedGateway = state =>
  selectGatewayById(state, selectSelectedGatewayId(state))

// Gateways.
const selectGtwsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectGtwsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)

export const selectGateways = createSelector(
  [selectGtwsIds, selectGatewayEntitiesStore],
  (ids, entities) => ids.map(id => entities[id]),
)
export const selectGatewaysTotalCount = state => selectGtwsTotalCount(state)

// Events.
export const selectGatewayEvents = createEventsSelector(ENTITY)
export const selectGatewayEventsError = createEventsErrorSelector(ENTITY)
export const selectGatewayEventsStatus = createEventsStatusSelector(ENTITY)
export const selectGatewayEventsInterrupted = createEventsInterruptedSelector(ENTITY)
export const selectGatewayEventsPaused = createEventsPausedSelector(ENTITY)
export const selectGatewayEventsTruncated = createEventsTruncatedSelector(ENTITY)
export const selectLatestGatewayEvent = createLatestEventSelector(ENTITY)
export const selectGatewayEventsFilter = createEventsFilterSelector(ENTITY)

// Rights.
export const selectGatewayRights = createRightsSelector(ENTITY)
export const selectGatewayPseudoRights = createPseudoRightsSelector(ENTITY)
export const selectGatewayRightsError = createErrorSelector(ENTITY)
export const selectGatewayRightsFetching = createFetchingSelector(GET_GTWS_RIGHTS_LIST_BASE)

// Statistics.
export const selectGatewayStatisticsConnectError = createErrorSelector(START_GTW_STATS_BASE)
export const selectGatewayStatisticsUpdateError = state => {
  const statistics = selectGatewayStatisticsStore(state) || {}

  return statistics.error
}
export const selectGatewayStatisticsError = state =>
  selectGatewayStatisticsConnectError(state) || selectGatewayStatisticsUpdateError(state)
export const selectGatewayStatisticsIsFetching = createFetchingSelector([
  START_GTW_STATS_BASE,
  UPDATE_GTW_STATS_BASE,
])
export const selectGatewayStatistics = state => {
  const statistics = selectGatewayStatisticsStore(state)

  return statistics.stats
}
