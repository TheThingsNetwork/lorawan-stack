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

export const GET_API_KEYS_LIST = 'GET_API_KEYS_LIST'
export const GET_API_KEYS_LIST_SUCCESS = 'GET_API_KEYS_LIST_SUCCESS'
export const GET_API_KEYS_LIST_FAILURE = 'GET_API_KEYS_LIST_FAILURE'

export const createGetApiKeysListActionType = name => (
  `GET_${name}_API_KEYS_LIST`
)

export const createGetApiKeysListSuccessActionType = name => (
  `GET_${name}_API_KEYS_LIST_SUCCESS`
)

export const createGetApiKeysListFailureActionType = name => (
  `GET_${name}_API_KEYS_LIST_FAILURE`
)

export const createGetApiKeyActionType = name => (
  `GET_${name}_API_KEY`
)

export const getApiKeysList = name => (id, params) => (
  { type: createGetApiKeysListActionType(name), id, params }
)

export const getApiKeysListSuccess = name => (id, keys, totalCount) => (
  { type: createGetApiKeysListSuccessActionType(name), id, keys, totalCount }
)

export const getApiKeysListFailure = name => (id, error) => (
  { type: createGetApiKeysListFailureActionType(name), id, error }
)
