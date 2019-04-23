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

export const GET_COLLABORATORS_LIST = 'GET_COLLABORATORS_LIST'
export const GET_COLLABORATORS_LIST_SUCCESS = 'GET_COLLABORATORS_LIST_SUCCESS'
export const GET_COLLABORATORS_LIST_FAILURE = 'GET_COLLABORATORS_LIST_FAILURE'

export const createGetCollaboratorsListActionType = name => (
  `GET_${name}_COLLABORATORS_LIST`
)

export const createGetCollaboratorsListSuccessActionType = name => (
  `GET_${name}_COLLABORATORS_LIST_SUCCESS`
)

export const createGetCollaboratorsListFailureActionType = name => (
  `GET_${name}_COLLABORATORS_LIST_FAILURE`
)

export const createGetCollaboratorActionType = name => (
  `GET_${name}_COLLABORATOR`
)

export const getCollaboratorsList = name => (id, filters) => (
  { type: createGetCollaboratorsListActionType(name), id, filters }
)

export const getCollaboratorsListSuccess = name => (id, collaborators, totalCount) => (
  { type: createGetCollaboratorsListSuccessActionType(name), id, collaborators, totalCount }
)

export const getCollaboratorsListFailure = name => (id, error) => (
  { type: createGetCollaboratorsListFailureActionType(name), id, error }
)

export const getCollaborator = name => id => (
  { type: createGetCollaboratorActionType(name), id }
)
