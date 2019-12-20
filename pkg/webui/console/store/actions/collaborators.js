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

import { createPaginationByParentRequestActions } from './pagination'
import { createRequestActions } from './lib'

export const SHARED_NAME = 'COLLABORATORS'

export const GET_COLLABORATOR_BASE = 'GET_COLLABORATOR'
export const [
  {
    request: GET_COLLABORATOR,
    success: GET_COLLABORATOR_SUCCESS,
    failure: GET_COLLABORATOR_FAILURE,
  },
  { request: getCollaborator, success: getCollaboratorSuccess, failure: getCollaboratorFailure },
] = createRequestActions(
  GET_COLLABORATOR_BASE,
  (parentType, parentId, collaboratorId, isUser) => ({
    parentType,
    parentId,
    collaboratorId,
    isUser,
  }),
  (parentType, parentId, collaboratorId, isUser, selector) => ({ selector }),
)

export const GET_COLLABORATORS_LIST_BASE = 'GET_COLLABORATORS_LIST'
export const [
  {
    request: GET_COLLABORATORS_LIST,
    success: GET_COLLABORATORS_LIST_SUCCESS,
    failure: GET_COLLABORATORS_LIST_FAILURE,
  },
  {
    request: getCollaboratorsList,
    success: getCollaboratorsListSuccess,
    failure: getCollaboratorsListFailure,
  },
] = createPaginationByParentRequestActions(SHARED_NAME)
