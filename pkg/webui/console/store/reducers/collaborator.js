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
  createGetCollaboratorActionType,
} from '../actions/collaborators'

import { createRequestActions } from '../actions/lib'

const defaultState = {
  collaborator: undefined,
}

const createNamedCollaboratorReducer = function (reducerName = '') {
  const GET_COLLABORATOR_BASE = createGetCollaboratorActionType(reducerName)
  const [{ success: GET_COLLABORATOR_SUCCESS }] = createRequestActions(GET_COLLABORATOR_BASE)

  return function (state = defaultState, { type, payload }) {
    switch (type) {
    case GET_COLLABORATOR_SUCCESS:
      return {
        ...state,
        collaborator: payload,
      }
    default:
      return state
    }
  }
}

export default createNamedCollaboratorReducer
