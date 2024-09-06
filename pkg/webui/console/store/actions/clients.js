// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationRequestActions,
  createPaginationDeleteActions,
  createPaginationRestoreActions,
  createPaginationBaseActionType,
} from '@ttn-lw/lib/store/actions/pagination'

export const CREATE_CLIENT_BASE = 'CREATE_CLIENT'
export const [
  { request: CREATE_CLIENT, success: CREATE_CLIENT_SUCCESS, failure: CREATE_CLIENT_FAILURE },
  { request: createClient, success: createClientSuccess, failure: createClientFailure },
] = createRequestActions(CREATE_CLIENT_BASE, (ownerId, client, isUserOwner) => ({
  ownerId,
  client,
  isUserOwner,
}))

export const GET_CLIENT_BASE = 'GET_CLIENT'
export const [
  { request: GET_CLIENT, success: GET_CLIENT_SUCCESS, failure: GET_CLIENT_FAILURE },
  { request: getClient, success: getClientSuccess, failure: getClientFailure },
] = createRequestActions(
  GET_CLIENT_BASE,
  id => ({ id }),
  (id, selector) => ({ selector }),
)

export const UPDATE_CLIENT_BASE = 'UPDATE_CLIENT'
export const [
  { request: UPDATE_CLIENT, success: UPDATE_CLIENT_SUCCESS, failure: UPDATE_CLIENT_FAILURE },
  { request: updateClient, success: updateClientSuccess, failure: updateClientFailure },
] = createRequestActions(UPDATE_CLIENT_BASE, (id, patch) => ({ id, patch }))

export const DELETE_CLIENT_BASE = 'DELETE_CLIENT'
export const [
  { request: DELETE_CLIENT, success: DELETE_CLIENT_SUCCESS, failure: DELETE_CLIENT_FAILURE },
  { request: deleteClient, success: deleteClientSuccess, failure: deleteClientFailure },
] = createPaginationDeleteActions('CLIENTS', (id, options) => ({ id, options }))

export const RESTORE_CLIENT_BASE = 'RESTORE_CLIENT'
export const [
  { request: RESTORE_CLIENT, success: RESTORE_CLIENT_SUCCESS, failure: RESTORE_CLIENT_FAILURE },
  { request: restoreClient, success: restoreClientSuccess, failure: restoreClientFailure },
] = createPaginationRestoreActions('CLIENTS', id => ({ id }))

export const GET_CLIENTS_LIST_BASE = createPaginationBaseActionType('CLIENTS')
export const [
  {
    request: GET_CLIENTS_LIST,
    success: GET_CLIENTS_LIST_SUCCESS,
    failure: GET_CLIENTS_LIST_FAILURE,
  },
  { request: getClientsList, success: getClientsSuccess, failure: getClientsFailure },
] = createPaginationRequestActions('CLIENTS')

export const GET_CLIENT_RIGHTS_BASE = 'GET_CLIENT_RIGHTS'
export const [
  {
    request: GET_CLIENT_RIGHTS,
    success: GET_CLIENT_RIGHTS_SUCCESS,
    failure: GET_CLIENT_RIGHTS_FAILURE,
  },
  { request: getClientRights, success: getClientRightsSuccess, failure: getClientRightsFailure },
] = createRequestActions(GET_CLIENT_RIGHTS_BASE, id => ({ id }))
