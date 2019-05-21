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
  createGetApiKeyActionType,
  createGetApiKeySuccessActionType,
  createGetApiKeyFailureActionType,
} from '../actions/api-key'

const defaultState = {
  fetching: false,
  error: undefined,
  key: undefined,
}

const createNamedApiKeyReducer = function (reducerName = '') {
  const GET_KEY = createGetApiKeyActionType(reducerName)
  const GET_KEY_SUCCESS = createGetApiKeySuccessActionType(reducerName)
  const GET_KEY_FAILURE = createGetApiKeyFailureActionType(reducerName)

  return function (state = defaultState, action) {
    switch (action.type) {
    case GET_KEY:
      return {
        fetching: true,
      }
    case GET_KEY_SUCCESS:
      return {
        ...state,
        key: action.key,
        fetching: false,
        error: undefined,
      }
    case GET_KEY_FAILURE:
      return {
        error: action.error,
        fetching: false,
      }
    default:
      return state
    }
  }
}

export default createNamedApiKeyReducer
