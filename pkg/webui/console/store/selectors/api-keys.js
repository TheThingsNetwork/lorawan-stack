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

import { GET_API_KEY_BASE, GET_API_KEYS_LIST_BASE } from '../actions/api-keys'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'

const ENTITY = 'apiKeys'

// Api Key
export const selectApiKeysStore = state => state.apiKeys || {}
export const selectApiKeysEntitiesStore = state => selectApiKeysStore(state).entities
export const selectApiKeyById = (state, id) => selectApiKeysEntitiesStore(state)[id]
export const selectSelectedApiKeyId = state => selectApiKeysStore(state).selectedApiKey
export const selectSelectedApiKey = state => selectApiKeyById(state, selectSelectedApiKeyId(state))
export const selectApiKeyFetching = createFetchingSelector(GET_API_KEY_BASE)
export const selectApiKeyError = createErrorSelector(GET_API_KEY_BASE)

// Api Keys
const createSelectApiKeysIdsSelector = createPaginationIdsSelectorByEntity(ENTITY)
const createSelectApiKeysTotalCountSelector = createPaginationTotalCountSelectorByEntity(ENTITY)
const createSelectApiKeysFetchingSelector = createFetchingSelector(GET_API_KEYS_LIST_BASE)
const createSelectApiKeysErrorSelector = createErrorSelector(GET_API_KEYS_LIST_BASE)

export const selectApiKeys = state =>
  createSelectApiKeysIdsSelector(state).map(id => selectApiKeyById(state, id))
export const selectApiKeysTotalCount = state => createSelectApiKeysTotalCountSelector(state)
export const selectApiKeysFetching = state => createSelectApiKeysFetchingSelector(state)
export const selectApiKeysError = state => createSelectApiKeysErrorSelector(state)
