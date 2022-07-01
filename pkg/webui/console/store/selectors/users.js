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

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'

import {
  GET_USERS_LIST_BASE,
  GET_USER_BASE,
  GET_USER_RIGHTS_LIST_BASE,
  GET_USER_INVITATIONS_BASE,
} from '@console/store/actions/users'

import { createRightsSelector, createPseudoRightsSelector } from './rights'

const ENTITY = 'users'

const selectUserStore = state => state.users

// User.
export const selectUserEntitiesStore = state => selectUserStore(state).entities
export const selectUserById = (state, id) => selectUserEntitiesStore(state)[id]
export const selectSelectedUserId = state => selectUserStore(state).selectedUser
export const selectSelectedUser = state => selectUserById(state, selectSelectedUserId(state))
export const selectUserFetching = createFetchingSelector(GET_USER_BASE)
export const selectUserError = createErrorSelector(GET_USER_BASE)

// Users.
const selectUsrsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectUsrsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectUsrsFetching = createFetchingSelector(GET_USERS_LIST_BASE)
const selectUsrsError = createErrorSelector(GET_USERS_LIST_BASE)

export const selectUsers = state => selectUsrsIds(state).map(id => selectUserById(state, id))
export const selectUsersTotalCount = state => selectUsrsTotalCount(state)
export const selectUsersFetching = state => selectUsrsFetching(state)
export const selectUsersError = state => selectUsrsError(state)

// Rights.
export const selectUserRights = createRightsSelector(ENTITY)
export const selectUserPseudoRights = createPseudoRightsSelector(ENTITY)
export const selectUserRightsError = createErrorSelector(GET_USER_RIGHTS_LIST_BASE)
export const selectUserRightsFetching = createFetchingSelector(GET_USER_RIGHTS_LIST_BASE)

// Invitations.
export const selectUserInvitations = state => selectUserStore(state).invitations
export const selectUserInvitationsTotalCount = state => selectUserStore(state).invitationsTotalCount
export const selectUserInvitationsFetching = createFetchingSelector(GET_USER_INVITATIONS_BASE)
export const selectUserInvitationsError = createErrorSelector(GET_USER_INVITATIONS_BASE)
