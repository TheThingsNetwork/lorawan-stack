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
  createGetApiKeysListActionType,
  createGetApiKeysListFailureActionType,
  createGetApiKeysListSuccessActionType,
} from '../actions/api-keys'

const defualtState = {
  fetching: false,
  keys: [],
  totalCount: 0,
  error: false,
}

const createNamedApiKeyReducer = function (reducerName = '') {
  const GET_LIST = createGetApiKeysListActionType(reducerName)
  const GET_LIST_SUCCESS = createGetApiKeysListSuccessActionType(reducerName)
  const GET_LIST_FAILURE = createGetApiKeysListFailureActionType(reducerName)

  return function (state = defualtState, action) {
    switch (action.type) {
    case GET_LIST:
      return {
        ...state,
        fetching: true,
      }
    case GET_LIST_FAILURE:
      return {
        ...state,
        fetching: false,
        keys: [],
        totalCount: 0,
        error: action.error,
      }
    case GET_LIST_SUCCESS:
      return {
        ...state,
        keys: action.keys,
        totalCount: action.totalCount,
        fetching: false,
      }
    default:
      return state
    }
  }
}

const createNamedAPIKeysReducer = function (reducerName = '') {
  const GET_LIST = createGetApiKeysListActionType(reducerName)
  const GET_LIST_SUCCESS = createGetApiKeysListSuccessActionType(reducerName)
  const GET_LIST_FAILURE = createGetApiKeysListFailureActionType(reducerName)
  const apiKey = createNamedApiKeyReducer(reducerName)

  return function (state = {}, action) {
    if (!action.id) {
      return state
    }

    switch (action.type) {
    case GET_LIST:
    case GET_LIST_FAILURE:
    case GET_LIST_SUCCESS:
      return {
        ...state,
        [action.id]: apiKey(state[action.id], action),
      }
    default:
      return state
    }
  }
}

export default createNamedAPIKeysReducer
