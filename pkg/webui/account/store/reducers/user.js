// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
  GET_USER_ME,
  GET_USER_ME_SUCCESS,
  GET_USER_ME_FAILURE,
  LOGOUT,
  LOGOUT_SUCCESS,
  LOGOUT_FAILURE,
} from '@account/store/actions/user'

const defaultState = {
  fetching: false,
  user: undefined,
  error: false,
}

const user = function(state = defaultState, { type, payload }) {
  switch (type) {
    case GET_USER_ME:
      return {
        ...state,
        fetching: true,
        user: undefined,
        error: false,
      }
    case GET_USER_ME_SUCCESS:
      return {
        ...state,
        fetching: false,
        user: payload,
        error: false,
      }
    case GET_USER_ME_FAILURE:
      return {
        ...state,
        fetching: false,
        user: undefined,
        error: payload,
      }
    case LOGOUT:
      return {
        ...state,
        fetching: true,
        error: false,
      }
    case LOGOUT_SUCCESS:
      return {
        ...state,
        user: undefined,
        fetching: false,
        error: false,
      }
    case LOGOUT_FAILURE:
      return {
        ...state,
        fetching: false,
        error: payload,
      }
    default:
      return state
  }
}

export default user
