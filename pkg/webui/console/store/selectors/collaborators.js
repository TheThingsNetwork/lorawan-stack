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

import { GET_COLLABORATOR_BASE, GET_COLLABORATORS_LIST_BASE } from '../actions/collaborators'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'

const ENTITY = 'collaborators'

// Collaborator
export const selectCollaboratorsStore = state => state.collaborators || {}
export const selectCollaboratorsEntitiesStore = state => selectCollaboratorsStore(state).entities
export const selectCollaboratorById = (state, id) => selectCollaboratorsEntitiesStore(state)[id]
export const selectSelectedCollaboratorId = state =>
  selectCollaboratorsStore(state).selectedCollaborator
export const selectSelectedCollaborator = state =>
  selectCollaboratorById(state, selectSelectedCollaboratorId(state))
export const selectCollaboratorFetching = createFetchingSelector(GET_COLLABORATOR_BASE)
export const selectCollaboratorError = createErrorSelector(GET_COLLABORATOR_BASE)
export const selectUserCollaborator = function(state) {
  const collaborator = selectSelectedCollaborator(state)

  if (collaborator && 'user_ids' in collaborator.ids) {
    return collaborator
  }
}
export const selectOrganizationCollaborator = function(state) {
  const collaborator = selectSelectedCollaborator(state)

  if (collaborator && 'organization_ids' in collaborator.ids) {
    return collaborator
  }
}

// Collaborators
const createSelectCollaboratorsIdsSelector = createPaginationIdsSelectorByEntity(ENTITY)
const createSelectCollaboratorsTotalCountSelector = createPaginationTotalCountSelectorByEntity(
  ENTITY,
)
const createSelectCollaboratorsFetchingSelector = createFetchingSelector(
  GET_COLLABORATORS_LIST_BASE,
)
const createSelectCollaboratorsErrorSelector = createErrorSelector(GET_COLLABORATORS_LIST_BASE)

export const selectCollaborators = state =>
  createSelectCollaboratorsIdsSelector(state).map(id => selectCollaboratorById(state, id))
export const selectCollaboratorsTotalCount = state =>
  createSelectCollaboratorsTotalCountSelector(state)
export const selectCollaboratorsFetching = state => createSelectCollaboratorsFetchingSelector(state)
export const selectCollaboratorsError = state => createSelectCollaboratorsErrorSelector(state)
