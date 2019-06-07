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

import { getApplicationId } from '../../../lib/selectors/id'
import { GET_APP_BASE, GET_APP_API_KEY_BASE } from '../actions/application'
import {
  eventsSelector,
  errorSelector as eventsErrorSelector,
  statusSelector as eventsStatusSelector,
} from './events'
import {
  rightsSelector,
  errorSelector as rightsErrorSelector,
  fetchingSelector as rightsFetchingSelector,
} from './rights'
import {
  apiKeysSelector,
  errorSelector as apiKeysErrorSelector,
  fetchingSelector as apiKeysFetchingSelector,
} from './api-keys'
import { apiKeySelector } from './api-key'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'applications'
const ENTITY_SINGLE = 'application'
const storeSelector = state => state.application

export const applicationSelector = state => storeSelector(state).application

export const fetchingSelector = state => storeSelector(state).fetching || false

export const errorSelector = state => storeSelector(state).error

const selectApplicationStore = state => state.application

export const selectSelectedApplicationId = function (state) {
  const application = selectApplicationStore(state)

  return getApplicationId(application)
}
export const selectSelectedApplication = state => selectApplicationStore(state)
export const selectApplicationFetching = createFetchingSelector(GET_APP_BASE)
export const selectApplicationError = createErrorSelector(GET_APP_BASE)

export const applicationEventsSelector = eventsSelector(ENTITY)

export const applicationEventsErrorSelector = eventsErrorSelector(ENTITY)

export const applicationEventsStatusSelector = eventsStatusSelector(ENTITY)

export const applicationRightsSelector = rightsSelector(ENTITY)

export const applicationRightsErrorSelector = rightsErrorSelector(ENTITY)

export const applicationRightsFetchingSelector = rightsFetchingSelector(ENTITY)

export const applicationKeysSelector = apiKeysSelector(ENTITY)

export const applicationKeysErrorSelector = apiKeysErrorSelector(ENTITY)

export const applicationKeysFetchingSelector = apiKeysFetchingSelector(ENTITY)

export const applicationKeySelector = apiKeySelector(ENTITY_SINGLE)

export const applicationKeyFetchingSelector = createFetchingSelector(GET_APP_API_KEY_BASE)

export const applicationKeyErrorSelector = createErrorSelector(GET_APP_API_KEY_BASE)

const selectLinkStore = state => state.link

export const selectApplicationLink = state => selectLinkStore(state).link

export const selectApplicationLinkStats = state => selectLinkStore(state).stats

export const selectApplicationLinkFetching = state => selectLinkStore(state).fetching || false

export const selectApplicationLinkError = state => selectLinkStore(state).error

export const selectApplicationLinkFormatters = function (state) {
  const link = selectApplicationLink(state) || {}

  return link.default_formatters
}

export const selectApplicationIsLinked = function (state) {
  const linkStore = selectLinkStore(state)
  const link = selectApplicationLink(state) || {}
  const error = selectApplicationLinkError(state)
  const stats = selectApplicationLinkStats(state)

  const hasBase = Boolean(link.api_key)
  const hasError = Boolean(error)
  const isLinked = linkStore.linked
  const hasStats = Boolean(stats)

  return hasBase && !hasError && isLinked && hasStats
}
