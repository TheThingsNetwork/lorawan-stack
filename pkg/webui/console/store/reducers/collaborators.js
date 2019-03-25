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
  createGetCollaboratorsListActionType,
  createGetCollaboratorsListFailureActionType,
  createGetCollaboratorsListSuccessActionType,
} from '../actions/collaborators'

const defaultState = {
  fetching: false,
  collaborators: [],
  totalCount: 0,
  error: false,
}

const createNamedCollaboratorReducer = function (reducerName = '') {
  const GET_LIST = createGetCollaboratorsListActionType(reducerName)
  const GET_LIST_SUCCESS = createGetCollaboratorsListSuccessActionType(reducerName)
  const GET_LIST_FAILURE = createGetCollaboratorsListFailureActionType(reducerName)

  return function (state = defaultState, action) {
    switch (action.type) {
    case GET_LIST:
      return {
        ...state,
        fetching: true,
        error: false,
      }
    case GET_LIST_FAILURE:
      return {
        ...state,
        error: action.error,
        fetching: false,
      }
    case GET_LIST_SUCCESS:
      return {
        ...state,
        error: false,
        fetching: false,
        collaborators: action.collaborators,
        totalCount: action.totalCount,
      }
    default:
      return state
    }
  }
}

const createNamedCollaboratorsReducer = function (reducerName = '') {
  const GET_LIST = createGetCollaboratorsListActionType(reducerName)
  const GET_LIST_SUCCESS = createGetCollaboratorsListSuccessActionType(reducerName)
  const GET_LIST_FAILURE = createGetCollaboratorsListFailureActionType(reducerName)
  const collaborators = createNamedCollaboratorReducer(reducerName)

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
        [action.id]: collaborators(state[action.id], action),
      }
    default:
      return state
    }
  }
}

export default createNamedCollaboratorsReducer
