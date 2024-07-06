// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import {
  GET_COLLABORATORS_LIST_SUCCESS,
  GET_COLLABORATOR_SUCCESS,
  GET_COLLABORATOR,
} from '@ttn-lw/lib/store/actions/collaborators'

const defaultState = {
  entities: {},
  selectedCollaborator: null,
  totalCount: null,
}

const collaborator = (state = {}, collaborator) => ({
  ...state,
  ...collaborator,
})

const collaborators = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_COLLABORATOR:
      return {
        ...state,
        selectedCollaborator: payload.collaboratorId,
      }
    case GET_COLLABORATOR_SUCCESS:
      const id = getCollaboratorId(payload)
      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: collaborator(state.entities[id], payload),
        },
      }
    case GET_COLLABORATORS_LIST_SUCCESS:
      return {
        ...state,
        entities: {
          ...payload.entities.reduce((acc, col) => {
            const id = getCollaboratorId(col)
            acc[id] = collaborator(state.entities[id], col)
            return acc
          }, {}),
        },
        totalCount: payload.totalCount,
      }
    default:
      return state
  }
}

export default collaborators
