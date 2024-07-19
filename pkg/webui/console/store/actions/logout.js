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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const GET_USER_ME_BASE = 'GET_USER_ME'
export const [
  { request: GET_USER_ME, success: GET_USER_ME_SUCCESS, failure: GET_USER_ME_FAILURE },
  { request: getUserMe, success: getUserMeSuccess, failure: getUserMeFailure },
] = createRequestActions(GET_USER_ME_BASE)

export const GET_USER_RIGHTS_BASE = 'GET_USER_RIGHTS'
export const [
  { request: GET_USER_RIGHTS, success: GET_USER_RIGHTS_SUCCESS, failure: GET_USER_RIGHTS_FAILURE },
  { request: getUserRights, success: getUserRightsSuccess, failure: getUserRightsFailure },
] = createRequestActions(GET_USER_RIGHTS_BASE)

export const APPLY_PERSISTED_STATE_BASE = 'APPLY_PERSISTED_STATE'
export const [
  {
    request: APPLY_PERSISTED_STATE,
    success: APPLY_PERSISTED_STATE_SUCCESS,
    failure: APPLY_PERSISTED_STATE_FAILURE,
  },
  {
    request: applyPersistedState,
    success: applyPersistedStateSuccess,
    failure: applyPersistedStateFailure,
  },
] = createRequestActions(APPLY_PERSISTED_STATE_BASE)

export const LOGOUT_BASE = 'LOGOUT'
export const [
  { request: LOGOUT, success: LOGOUT_SUCCESS, failure: LOGOUT_FAILURE },
  { request: logout, success: logoutSuccess, failure: logoutFailure },
] = createRequestActions(LOGOUT_BASE)
