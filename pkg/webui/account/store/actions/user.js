// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

export const GET_USER_BASE = 'GET_USER'
export const [
  { request: GET_USER, success: GET_USER_SUCCESS, failure: GET_USER_FAILURE },
  { request: getUser, success: getUserSuccess, failure: getUserFailure },
] = createRequestActions(
  GET_USER_BASE,
  id => ({ id }),
  (id, selector) => ({ selector }),
)

export const UPDATE_USER_BASE = 'UPDATE_USER'
export const [
  { request: UPDATE_USER, success: UPDATE_USER_SUCCESS, failure: UPDATE_USER_FAILURE },
  { request: updateUser, success: updateUserSuccess, failure: updateUserFailure },
] = createRequestActions(UPDATE_USER_BASE, ({ id, patch }) => ({ id, patch }))

export const DELETE_USER_BASE = 'DELETE_USER'
export const [
  { request: DELETE_USER, success: DELETE_USER_SUCCESS, failure: DELETE_USER_FAILURE },
  { request: deleteUser, success: deleteUserSuccess, failure: deleteUserFailure },
] = createRequestActions(DELETE_USER_BASE, id => ({ id }))

export const LOGOUT_BASE = 'LOGOUT'
export const [
  { request: LOGOUT, success: LOGOUT_SUCCESS, failure: LOGOUT_FAILURE },
  { request: logout, success: logoutSuccess, failure: logoutFailure },
] = createRequestActions(LOGOUT_BASE)
