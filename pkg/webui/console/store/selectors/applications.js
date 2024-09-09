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

import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import { GET_APP_LINK_BASE } from '@console/store/actions/link'
import {
  GET_APPS_RIGHTS_LIST_BASE,
  GET_APP_BASE,
  GET_APP_DEV_COUNT_BASE,
} from '@console/store/actions/applications'

import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
  createEventsInterruptedSelector,
  createEventsPausedSelector,
  createEventsTruncatedSelector,
  createEventsFilterSelector,
} from './events'
import { createRightsSelector, createPseudoRightsSelector } from './rights'

const ENTITY = 'applications'

// Application.
export const selectApplicationStore = state => state.applications
export const selectApplicationEntitiesStore = state => selectApplicationStore(state).entities
export const selectApplicationDerivedStore = state => selectApplicationStore(state).derived
export const selectApplicationCombinedStore = createSelector(
  [selectApplicationEntitiesStore, selectApplicationDerivedStore],
  (entities, derived) =>
    Object.keys(entities).reduce((acc, id) => {
      acc[id] = {
        ...entities[id],
        ...derived[id],
      }
      return acc
    }, {}),
)
export const selectApplicationDerivedById = (state, id) => selectApplicationDerivedStore(state)[id]
export const selectApplicationById = (state, id) => selectApplicationEntitiesStore(state)[id]
export const selectSelectedApplicationId = state =>
  selectApplicationStore(state).selectedApplication
export const selectApplicationNameById = (state, id) =>
  (selectApplicationById(state, id) || {}).name
export const selectSelectedApplication = state =>
  selectApplicationById(state, selectSelectedApplicationId(state))
export const selectApplicationFetching = createFetchingSelector(GET_APP_BASE)
export const selectApplicationError = createErrorSelector(GET_APP_BASE)
export const selectApplicationDeviceCount = (state, id) =>
  selectApplicationDerivedById(state, id)?.deviceCount
export const selectApplicationDeviceCountFetching = createFetchingSelector(GET_APP_DEV_COUNT_BASE)
export const selectApplicationDeviceCountError = createErrorSelector(GET_APP_DEV_COUNT_BASE)
export const selectApplicationDerivedLastSeen = (state, id) =>
  (selectApplicationDerivedById(state, id) || {}).lastSeen
export const selectApplicationDevEUICount = state =>
  selectApplicationById(state, selectSelectedApplicationId(state)).dev_eui_counter || 0

// Applications.
const selectAppsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectAppsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)

export const selectApplications = createSelector(
  [selectAppsIds, selectApplicationEntitiesStore],
  (ids, entities) => ids.map(id => entities[id]),
)
export const selectApplicationsTotalCount = state => selectAppsTotalCount(state)
export const selectApplicationsWithDeviceCounts = createSelector(
  [selectApplications, selectApplicationDerivedStore],
  (applications, derived) =>
    applications.map(app => ({
      ...app,
      _devices: derived[app.ids.application_id]?.deviceCount,
    })),
)

// Events.
export const selectApplicationEvents = createEventsSelector(ENTITY)
export const selectApplicationEventsError = createEventsErrorSelector(ENTITY)
export const selectApplicationEventsStatus = createEventsStatusSelector(ENTITY)
export const selectApplicationEventsInterrupted = createEventsInterruptedSelector(ENTITY)
export const selectApplicationEventsPaused = createEventsPausedSelector(ENTITY)
export const selectApplicationEventsTruncated = createEventsTruncatedSelector(ENTITY)
export const selectApplicationEventsFilter = createEventsFilterSelector(ENTITY)

// Rights.
export const selectApplicationRights = createRightsSelector(ENTITY)
export const selectApplicationPseudoRights = createPseudoRightsSelector(ENTITY)
export const selectApplicationRightsError = createErrorSelector(GET_APPS_RIGHTS_LIST_BASE)

// Link.
const selectLinkStore = state => state.link
export const selectApplicationLink = state => selectLinkStore(state).link
export const selectApplicationLinkError = createErrorSelector(GET_APP_LINK_BASE)
export const selectApplicationLinkFormatters = state => {
  const link = selectApplicationLink(state) || {}

  return link.default_formatters
}

export const selectApplicationLinkSkipPayloadCrypto = state => {
  const link = selectApplicationLink(state) || {}

  return link.skip_payload_crypto || false
}

export const selectMqttConnectionInfo = state => selectApplicationStore(state).mqtt
