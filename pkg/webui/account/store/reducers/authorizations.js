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

import {
  GET_AUTHORIZATIONS_LIST_SUCCESS,
  GET_ACCESS_TOKENS_LIST_SUCCESS,
  DELETE_ALL_TOKENS_SUCCESS,
  DELETE_ACCESS_TOKEN_SUCCESS,
} from '@account/store/actions/authorizations'

const defaultState = {
  authorizations: [],
  authorizationsTotalCount: undefined,
  tokens: [],
  tokensTotalCount: undefined,
}

const authorizations = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_AUTHORIZATIONS_LIST_SUCCESS:
      return {
        ...state,
        authorizations: payload.entities,
        authorizationsTotalCount: payload.authorizationsTotalCount,
      }
    case GET_ACCESS_TOKENS_LIST_SUCCESS:
      return {
        ...state,
        tokens: payload.entities,
        tokensTotalCount: payload.tokensTotalCount,
      }
    case DELETE_ALL_TOKENS_SUCCESS:
      return {
        ...state,
        tokens: [],
        tokensTotalCount: 0,
      }
    case DELETE_ACCESS_TOKEN_SUCCESS:
      return {
        ...state,
        tokens: state.tokens.filter(token => token.id !== payload.id),
        tokensTotalCount: state.tokensTotalCount - 1,
      }
    default:
      return state
  }
}

export default authorizations
