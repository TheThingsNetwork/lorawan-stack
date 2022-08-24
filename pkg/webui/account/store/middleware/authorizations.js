// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import tts from '@account/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as authorizations from '@account/store/actions/authorizations'

const getAuthorizationsLogic = createRequestLogic({
  type: authorizations.GET_AUTHORIZATIONS_LIST,
  process: async ({ action }) => {
    const {
      id,
      params: { page, limit, order },
    } = action.payload

    const res = await tts.Authorizations.getAllAuthorizations(id, { page, limit, order })

    return { entities: res.authorizations, authorizationsTotalCount: res.totalCount }
  },
})

const deleteAuthorizationLogic = createRequestLogic({
  type: authorizations.DELETE_AUTHORIZATION,
  process: async ({ action }) => {
    const { id, targetId } = action.payload

    return await tts.Authorizations.deleteAuthorization(id, targetId)
  },
})

const getAccessTokensLogic = createRequestLogic({
  type: authorizations.GET_ACCESS_TOKENS_LIST,
  process: async ({ action }) => {
    const {
      parentId,
      parentType,
      params: { page, limit, order },
    } = action.payload
    const res = await tts.Authorizations.getAllTokens(parentType, parentId, { page, limit, order })

    return { entities: res.tokens, tokensTotalCount: res.totalCount }
  },
})

const deleteAccessTokenLogic = createRequestLogic({
  type: authorizations.DELETE_ACCESS_TOKEN,
  process: async ({ action }) => {
    const { routeParams, id } = action.payload

    return await tts.Authorizations.deleteToken(...routeParams, id)
  },
})

const deleteAllTokensLogic = createRequestLogic({
  type: authorizations.DELETE_ALL_TOKENS,
  process: async ({ action }) => {
    const { routeParams, id } = action.payload

    return await Promise.all(id.map(id => tts.Authorizations.deleteToken(...routeParams, id)))
  },
})

export default [
  getAuthorizationsLogic,
  deleteAuthorizationLogic,
  getAccessTokensLogic,
  deleteAccessTokenLogic,
  deleteAllTokensLogic,
]
