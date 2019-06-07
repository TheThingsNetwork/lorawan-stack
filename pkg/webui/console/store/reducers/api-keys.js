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

import { createGetApiKeysListActionType } from '../actions/api-keys'
import { createRequestActions } from '../actions/lib'

const defaultState = {
  keys: [],
  totalCount: 0,
}

const createNamedApiKeyReducer = function (reducerName = '') {
  const GET_LIST_BASE = createGetApiKeysListActionType(reducerName)
  const [{ success: GET_LIST_SUCCESS }] = createRequestActions(GET_LIST_BASE)

  return function (state = defaultState, { type, payload }) {
    switch (type) {
    case GET_LIST_SUCCESS:
      return {
        ...state,
        keys: payload.api_keys,
        totalCount: payload.totalCount,
      }
    default:
      return state
    }
  }
}

const createNamedAPIKeysReducer = function (reducerName = '') {
  const GET_LIST_BASE = createGetApiKeysListActionType(reducerName)
  const [{ success: GET_LIST_SUCCESS }] = createRequestActions(GET_LIST_BASE)
  const apiKey = createNamedApiKeyReducer(reducerName)

  return function (state = {}, action) {
    const { payload } = action
    if (!payload || !payload.id) {
      return state
    }

    switch (action.type) {
    case GET_LIST_SUCCESS:
      return {
        ...state,
        [payload.id]: apiKey(state[payload.id], action),
      }
    default:
      return state
    }
  }
}

export default createNamedAPIKeysReducer
