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

import { createPaginationByParentRequestActions } from './pagination'
import { createRequestActions } from './lib'

export const SHARED_NAME = 'API_KEYS'

export const GET_API_KEY_BASE = 'GET_API_KEY'
export const [
  { request: GET_API_KEY, success: GET_API_KEY_SUCCESS, failure: GET_API_KEY_FAILURE },
  { request: getApiKey, success: getApiKeySuccess, failure: getApiKeyFailure },
] = createRequestActions(
  GET_API_KEY_BASE,
  (parentType, parentId, keyId) => ({ parentType, parentId, keyId }),
  (parentType, parentId, keyId, selector) => ({ selector }),
)

export const GET_API_KEYS_LIST_BASE = 'GET_API_KEYS_LIST'
export const [
  {
    request: GET_API_KEYS_LIST,
    success: GET_API_KEYS_LIST_SUCCESS,
    failure: GET_API_KEYS_LIST_FAILURE,
  },
  { request: getApiKeysList, success: getApiKeysListSuccess, failure: getApiKeysListFailure },
] = createPaginationByParentRequestActions(SHARED_NAME)
