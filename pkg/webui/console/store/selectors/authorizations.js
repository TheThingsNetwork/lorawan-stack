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

import { createSelector } from 'reselect'

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import {
  GET_AUTHORIZATIONS_LIST_BASE,
  GET_ACCESS_TOKENS_LIST_BASE,
} from '@console/store/actions/authorizations'

const selectAuthorizationsStore = state => state.authorizations

export const selectAuthorizations = state => selectAuthorizationsStore(state).authorizations

export const selectSelectedAuthorization = createSelector(
  [selectAuthorizationsStore, (_, clientId) => clientId],
  (authorizationsStore, clientId) => {
    const authorizations = authorizationsStore.authorizations.reduce((acc, cur) => {
      const id = cur.client_ids.client_id
      acc[id] = cur
      return acc
    }, {})
    return authorizations[clientId]
  },
)
export const selectAuthorizationsTotalCount = state =>
  selectAuthorizationsStore(state).authorizationsTotalCount
export const selectAuthorizationsFetching = createFetchingSelector(GET_AUTHORIZATIONS_LIST_BASE)
export const selectAuthorizationsError = createErrorSelector(GET_AUTHORIZATIONS_LIST_BASE)

export const selectTokens = state => selectAuthorizationsStore(state).tokens
export const selectTokenIds = createSelector([selectTokens], tokens =>
  tokens.map(token => token.id),
)
export const selectTokensTotalCount = state => selectAuthorizationsStore(state).tokensTotalCount
export const selectTokensFetching = createFetchingSelector(GET_ACCESS_TOKENS_LIST_BASE)
export const selectTokensError = createErrorSelector(GET_ACCESS_TOKENS_LIST_BASE)
