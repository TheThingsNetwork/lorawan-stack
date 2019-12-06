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

import { createPaginationRequestActions, createPaginationBaseActionType } from './pagination'
import { createRequestActions } from './lib'

export const SHARED_NAME = 'USER'

export const GET_USER_BASE = 'GET_USER'
export const [
  { request: GET_USER, success: GET_USER_SUCCESS, failure: GET_USER_FAILURE },
  { request: getUser, success: getUserSuccess, failure: getUserFailure },
] = createRequestActions(GET_USER_BASE, id => ({ id }), (id, selector) => ({ selector }))

export const GET_USERS_LIST_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  { request: GET_USERS_LIST, success: GET_USERS_LIST_SUCCESS, failure: GET_USERS_LIST_FAILURE },
  { request: getUsersList, success: getUsersSuccess, failure: getUsersFailure },
] = createPaginationRequestActions(SHARED_NAME)
