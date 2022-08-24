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

import autoBind from 'auto-bind'

import Marshaler from '../util/marshaler'

class Authorizations {
  constructor(registry) {
    this._api = registry

    autoBind(this)
  }

  async getAllAuthorizations(userId, params) {
    const result = await this._api.OAuthAuthorizationRegistry.List(
      {
        routeParams: { 'user_ids.user_id': userId },
      },
      { ...params },
    )

    return Marshaler.payloadListResponse('authorizations', result)
  }

  async deleteAuthorization(userId, client_id) {
    const result = await this._api.OAuthAuthorizationRegistry.Delete({
      routeParams: { 'user_ids.user_id': userId, 'client_ids.client_id': client_id },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getAllTokens(userId, client_id, params) {
    const result = await this._api.OAuthAuthorizationRegistry.ListTokens(
      {
        routeParams: { 'user_ids.user_id': userId, 'client_ids.client_id': client_id },
      },
      { ...params },
    )

    return Marshaler.payloadListResponse('tokens', result)
  }

  async deleteToken(userId, client_id, id) {
    const result = await this._api.OAuthAuthorizationRegistry.DeleteToken({
      routeParams: { 'user_ids.user_id': userId, 'client_ids.client_id': client_id, id },
    })

    return Marshaler.payloadSingleResponse(result)
  }
}

export default Authorizations
