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

import { INITIALIZE_SUCCESS } from '@ttn-lw/lib/store/actions/init'

import {
  GET_USER,
  GET_USER_SUCCESS,
  GET_USER_FAILURE,
  UPDATE_USER_SUCCESS,
} from '@account/store/actions/user'

const defaultState = {
  user: undefined,
  sessionId: undefined,
}

const user = (state = defaultState, { type, payload }) => {
  switch (type) {
    case INITIALIZE_SUCCESS:
      if (typeof payload !== 'string') {
        return state
      }

      return {
        ...state,
        sessionId: payload,
      }
    case GET_USER:
      return {
        ...state,
        user: undefined,
      }
    case GET_USER_SUCCESS:
      return {
        ...state,
        user: payload,
      }
    case GET_USER_FAILURE:
      return {
        ...state,
        user: undefined,
      }
    case UPDATE_USER_SUCCESS:
      return {
        ...state,
        user: {
          ...state.user,
          ...payload,
        },
      }
    default:
      return state
  }
}

export default user
