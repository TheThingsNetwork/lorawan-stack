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
  eventsSelector,
  errorSelector as eventsErrorSelector,
  statusSelector as eventsStatusSelector,
} from './events'

import {
  apiKeysStoreSelector,
  fetchingSelector as apiKeysFetchingSelector,
  errorSelector as apiKeysErrorSelector,
  totalCountSelector as apiKeysTotalCountSelector,
} from './api-keys'

const ENTITY = 'gateways'

const storeSelector = state => state.gateway

export const gatewaySelector = state => storeSelector(state).gateway

export const fetchingSelector = function (state) {
  const store = storeSelector(state)

  return store.fetching || false
}

export const errorSelector = function (state) {
  const store = storeSelector(state)

  return store.error
}

const statisticsStoreSelector = function (state) {
  const store = storeSelector(state)

  return store.statistics
}

export const statisticsSelector = function (state) {
  const store = statisticsStoreSelector(state)

  return store.stats
}

export const statisticsErrorSelector = function (state) {
  const store = statisticsStoreSelector(state)

  return store.error
}

export const statisticsIsAvailableSelector = function (state) {
  const store = statisticsStoreSelector(state)

  return store.available
}

export const statisticsIsFetchingSelector = function (state) {
  const store = statisticsStoreSelector(state)

  return store.fetching
}

export const gatewayEventsSelector = eventsSelector(ENTITY)

export const gatewayEventsErrorSelector = eventsErrorSelector(ENTITY)

export const gatewayEventsStatusSelector = eventsStatusSelector(ENTITY)

export const gatewayApiKeysStoreSelector = apiKeysStoreSelector(ENTITY)

export const gatewayTotalCountSelector = apiKeysTotalCountSelector(ENTITY)

export const gatewayErrorSelector = apiKeysErrorSelector(ENTITY)

export const gatewayFetchingSelector = apiKeysFetchingSelector(ENTITY)
