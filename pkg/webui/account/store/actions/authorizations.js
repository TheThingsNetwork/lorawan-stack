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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const GET_AUTHORIZATIONS_LIST_BASE = 'GET_AUTHORIZATIONS_LIST'
export const [
  {
    request: GET_AUTHORIZATIONS_LIST,
    success: GET_AUTHORIZATIONS_LIST_SUCCESS,
    failure: GET_AUTHORIZATIONS_LIST_FAILURE,
  },
  {
    request: getAuthorizationsList,
    success: getAuthorizationsSuccess,
    failure: getAuthorizationsFailure,
  },
] = createRequestActions(
  GET_AUTHORIZATIONS_LIST_BASE,
  userId => ({ userId }),
  (userId, selector) => ({ selector }),
)

export const GET_AUTHORIZATION_TOKENS_LIST_BASE = 'GET_AUTHORIZATION_TOKENS_LIST'
export const [
  {
    request: GET_AUTHORIZATION_TOKENS_LIST,
    success: GET_AUTHORIZATION_TOKENS_LIST_SUCCESS,
    failure: GET_AUTHORIZATION_TOKENS_LIST_FAILURE,
  },
  {
    request: getAuthorizationTokensList,
    success: getAuthorizationTokensSuccess,
    failure: getAuthorizationTokensFailure,
  },
] = createRequestActions(
  GET_AUTHORIZATION_TOKENS_LIST_BASE,
  (userId, clientId) => ({ userId, clientId }),
  (userId, clientId, selector) => ({ selector }),
)

export const DELETE_AUTHORIZATION_BASE = 'DELETE_AUTHORIZATION'
export const [
  {
    request: DELETE_AUTHORIZATION,
    success: DELETE_AUTHORIZATION_SUCCESS,
    failure: DELETE_AUTHORIZATION_FAILURE,
  },
  {
    request: deleteAuthorization,
    success: deleteAuthorizationSuccess,
    failure: deleteAuthorizationFailure,
  },
] = createRequestActions(
  DELETE_AUTHORIZATION_BASE,
  (userId, clientId) => ({ userId, clientId }),
  (userId, clientId, selector) => ({ selector }),
)

export const DELETE_AUTHORIZATION_TOKEN_BASE = 'DELETE_AUTHORIZATION_TOKEN'
export const [
  {
    request: DELETE_AUTHORIZATION_TOKEN,
    success: DELETE_AUTHORIZATION_TOKEN_SUCCESS,
    failure: DELETE_AUTHORIZATION_TOKEN_FAILURE,
  },
  {
    request: deleteAuthorizationToken,
    success: deleteAuthorizationTokenSuccess,
    failure: deleteAuthorizationTokenFailure,
  },
] = createRequestActions(DELETE_AUTHORIZATION_TOKEN_BASE, (userId, clientId, id) => ({
  userId,
  clientId,
  id,
}))

export const DELETE_ALL_TOKENS_BASE = 'DELETE_ALL_TOKENS'
export const [
  {
    request: DELETE_ALL_TOKENS,
    success: DELETE_ALL_TOKENS_SUCCESS,
    failure: DELETE_ALL_TOKENS_FAILURE,
  },
  { request: deleteAllTokens, success: deleteAllTokensSuccess, failure: deleteAllTokensFailure },
] = createRequestActions(DELETE_ALL_TOKENS_BASE, (userId, clientId, ids) => ({
  userId,
  clientId,
  ids,
}))
