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

import { GET_USER_ME_BASE, LOGOUT_BASE } from '../actions/user'

import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const selectUserStore = state => state.user

export const selectUser = state => selectUserStore(state).user
export const selectUserError = createErrorSelector([GET_USER_ME_BASE, LOGOUT_BASE])
export const selectUserFetching = createFetchingSelector(GET_USER_ME_BASE)

export const selectUserId = function(state) {
  const user = selectUser(state)

  if (!Boolean(user)) {
    return undefined
  }

  return user.ids.user_id
}

export const selectUserName = state => {
  const user = selectUser(state)

  return user ? user.name : undefined
}

export const selectUserNameOrId = state => {
  const user = selectUser(state)

  return user ? user.name || user.ids.user_id : undefined
}

export const selectUserIsAdmin = function(state) {
  const user = selectUser(state)

  return user ? user.isAdmin : false
}

export const selectUserRights = state => selectUserStore(state).rights
