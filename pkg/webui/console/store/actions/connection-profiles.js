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
import { createPaginationRequestActions } from '@ttn-lw/lib/store/actions/pagination'

export const GET_CONNECTION_PROFILES_LIST_BASE = 'GET_CONNECTION_PROFILES_LIST'
export const [
  {
    request: GET_CONNECTION_PROFILES_LIST,
    success: GET_CONNECTION_PROFILES_LIST_SUCCESS,
    failure: GET_CONNECTION_PROFILES_LIST_FAILURE,
  },
  {
    request: getConnectionProfilesList,
    success: getConnectionProfilesListSuccess,
    failure: getConnectionProfilesListFailure,
  },
] = createPaginationRequestActions(
  GET_CONNECTION_PROFILES_LIST_BASE,
  ({ page, limit, entityId, type } = {}) => ({
    params: { page, limit },
    entityId,
    type,
  }),
)

export const [
  {
    request: CREATE_CONNECTION_PROFILE,
    success: CREATE_CONNECTION_PROFILE_SUCCESS,
    failure: CREATE_CONNECTION_PROFILE_FAILURE,
  },
  {
    request: createConnectionProfile,
    success: createConnectionProfileSuccess,
    failure: createConnectionProfileFailure,
  },
] = createRequestActions(`CREATE_CONNECTION_PROFILE`, profile => ({
  profile,
}))

export const DELETE_CONNECTION_PROFILE_BASE = 'DELETE_CONNECTION_PROFILE'
export const [
  {
    request: DELETE_CONNECTION_PROFILE,
    success: DELETE_CONNECTION_PROFILE_SUCCESS,
    failure: DELETE_CONNECTION_PROFILE_FAILURE,
  },
  {
    request: deleteConnectionProfile,
    success: deleteConnectionProfileSuccess,
    failure: deleteConnectionProfileFailure,
  },
] = createRequestActions(DELETE_CONNECTION_PROFILE_BASE, (id, entityId, type) => ({
  type,
  entityId,
  id,
}))

export const [
  {
    request: UPDATE_CONNECTION_PROFILE,
    success: UPDATE_CONNECTION_PROFILE_SUCCESS,
    failure: UPDATE_CONNECTION_PROFILE_FAILURE,
  },
  {
    request: updateConnectionProfile,
    success: updateConnectionProfileSuccess,
    failure: updateConnectionProfileFailure,
  },
] = createRequestActions(
  `UPDATE_CONNECTION_PROFILE`,
  (id, patch) => ({ id, patch }),
  (id, patch, selector) => ({ selector }),
)

export const [
  {
    request: GET_ACCESS_POINTS,
    success: GET_ACCESS_POINTS_SUCCESS,
    failure: GET_ACCESS_POINTS_FAILURE,
  },
  { request: getAccessPoints, success: getAccessPointsSuccess, failure: getAccessPointsFailure },
] = createRequestActions('GET_ACCESS_POINTS', (gatewayId, gatewayEui) => ({
  gatewayId,
  gatewayEui,
}))
