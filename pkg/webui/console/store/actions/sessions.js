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

export const GET_USER_SESSIONS_LIST_BASE = 'GET_USER_SESSIONS_LIST'
export const [
  {
    request: GET_USER_SESSIONS_LIST,
    success: GET_USER_SESSIONS_LIST_SUCCESS,
    failure: GET_USER_SESSIONS_LIST_FAILURE,
  },
  {
    request: getUserSessionsList,
    success: getUserSessionsListSuccess,
    failure: getUserSessionsListFailure,
  },
] = createRequestActions(
  GET_USER_SESSIONS_LIST_BASE,
  (id, params) => ({ id, params }),
  (id, params, selector) => ({ selector }),
)

export const DELETE_USER_SESSION_BASE = 'DELETE_USER_SESSION'
export const [
  {
    request: DELETE_USER_SESSION,
    success: DELETE_USER_SESSION_SUCCESS,
    failure: DELETE_USER_SESSION_FAILURE,
  },
  {
    request: deleteUserSession,
    success: deleteUserSessionSuccess,
    failure: deleteUserSessionFailure,
  },
] = createRequestActions(DELETE_USER_SESSION_BASE, (user, sessionId) => ({ user, sessionId }))

export const GET_ACTIVE_USER_SESSION_ID_BASE = 'GET_ACTIVE_USER_SESSION_ID'
export const [
  { success: GET_ACTIVE_USER_SESSION_ID_SUCCESS },
  { success: getActiveUserSessionIdSuccess },
] = createRequestActions(GET_ACTIVE_USER_SESSION_ID_BASE)
