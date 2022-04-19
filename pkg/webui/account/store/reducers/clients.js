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
  GET_CLIENT,
  GET_CLIENT_SUCCESS,
  GET_CLIENTS_LIST_SUCCESS,
  UPDATE_CLIENT_SUCCESS,
  GET_CLIENT_RIGHTS_SUCCESS,
} from '@account/store/actions/clients'

const defaultState = {
  entities: {},
  totalCount: null,
  selectedClient: null,
  rights: {
    regular: [],
    pseudo: [],
  },
}

const client = (state = {}, client) => ({
  ...state,
  ...client,
})

const clients = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_CLIENT:
      return {
        ...state,
        selectedClient: payload.id,
      }
    case GET_CLIENTS_LIST_SUCCESS:
      const clients = payload.entities.reduce(
        (acc, c) => {
          const id = c.ids.client_id

          acc[id] = client(acc[id], c)
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities: clients,
        totalCount: payload.totalCount,
      }
    case GET_CLIENT_SUCCESS:
    case UPDATE_CLIENT_SUCCESS:
      const id = payload.ids.client_id

      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: client(state.entities[id], payload),
        },
      }
    case GET_CLIENT_RIGHTS_SUCCESS:
      return {
        ...state,
        rights: {
          regular: payload.filter(right => !right.endsWith('_ALL')),
          pseudo: payload.filter(right => right.endsWith('_ALL')),
        },
      }
    default:
      return state
  }
}

export default clients
