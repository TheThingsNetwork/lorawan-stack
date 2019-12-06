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

import { getUserId } from '../../../lib/selectors/id'
import {
  GET_USERS_LIST_SUCCESS,
  GET_USER_SUCCESS,
  UPDATE_USER_SUCCESS,
  GET_USER,
} from '../actions/users'

const initialState = {
  entities: {},
  selectedUser: null,
}

const user = function(state = {}, user) {
  return {
    ...state,
    ...user,
  }
}

const users = function(state = initialState, { type, payload, meta }) {
  switch (type) {
    case GET_USER:
      return {
        ...state,
        selectedUser: payload.id,
      }
    case UPDATE_USER_SUCCESS:
    case GET_USER_SUCCESS:
      const id = getUserId(payload)

      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: user(state.entities[id], payload),
        },
      }
    case GET_USERS_LIST_SUCCESS:
      const entities = payload.entities.reduce(
        function(acc, app) {
          const id = getUserId(app)

          acc[id] = user(acc[id], app)
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities,
      }
    default:
      return state
  }
}

export default users
