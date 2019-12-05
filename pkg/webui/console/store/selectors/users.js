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

import { GET_USERS_LIST_BASE } from '../actions/users'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'

const ENTITY = 'users'

const selectUserStore = state => state.users

export const selectUserEntitiesStore = state => selectUserStore(state).entities
export const selectUserById = (state, id) => selectUserEntitiesStore(state)[id]
export const selectSelectedUserId = state => selectUserStore(state).selectedUsers
export const selectSelectedUser = state => selectUserById(state, selectSelectedUserId(state))

// Users
const selectUsrsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectUsrsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectUsrsFetching = createFetchingSelector(GET_USERS_LIST_BASE)
const selectUsrsError = createErrorSelector(GET_USERS_LIST_BASE)

export const selectUsers = state => selectUsrsIds(state).map(id => selectUserById(state, id))
export const selectUsersTotalCount = state => selectUsrsTotalCount(state)
export const selectUsersFetching = state => selectUsrsFetching(state)
export const selectUsersError = state => selectUsrsError(state)
