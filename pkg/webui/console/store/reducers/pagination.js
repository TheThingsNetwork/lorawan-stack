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
  createPaginationByIdRequestActions,
  createPaginationRequestActions,
  createPaginationDeleteActions,
  createPaginationByIdDeleteActions,
} from '../actions/pagination'

const defaultState = {
  ids: [],
  totalCount: undefined,
}

export const createNamedPaginationReducer = function (reducerName, entityIdSelector) {
  const [{ success: GET_PAGINATION_SUCCESS }] = createPaginationRequestActions(reducerName)
  const [{ success: DELETE_PAGINATION_SUCCESS }] = createPaginationDeleteActions(reducerName)

  return function (state = defaultState, { type, payload }) {
    switch (type) {
    case GET_PAGINATION_SUCCESS:
      return {
        ...state,
        totalCount: payload.totalCount,
        ids: payload.entities.map(entityIdSelector),
      }
    case DELETE_PAGINATION_SUCCESS:
      return {
        ...state,
        totalCount: state.totalCount - 1,
        ids: state.ids.filter(id => id !== payload.id),
      }
    default:
      return state
    }
  }
}

export const createNamedPaginationReducerById = function (reducerName, entityIdSelector) {
  const [{ success: GET_PAGINATION_SUCCESS }] = createPaginationByIdRequestActions(reducerName)
  const [{ success: DELETE_PAGINATION_SUCCESS }] = createPaginationByIdDeleteActions(reducerName)
  const [ , { success }] = createPaginationDeleteActions(reducerName)
  const paginationReducer = createNamedPaginationReducer(reducerName, entityIdSelector)

  return function (state = {}, action) {
    const { id } = action.payload

    if (!id) {
      return state
    }

    switch (action.type) {
    case GET_PAGINATION_SUCCESS:
      return {
        ...state,
        [id]: paginationReducer(state[id], action),
      }
    case DELETE_PAGINATION_SUCCESS:
      return {
        ...state,
        [id]: paginationReducer(state[id], success({ id: action.payload.targetId })),
      }
    default:
      return state
    }
  }
}
