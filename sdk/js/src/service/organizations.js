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

import Marshaler from '../util/marshaler'

class Organizations {
  constructor(api) {
    this._api = api
  }

  // Retrieval

  async getAll(params, selector) {
    const response = await this._api.OrganizationRegistry.List(undefined, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return Marshaler.payloadListResponse('organizations', response)
  }

  // Create

  async create(userId, organization) {
    const response = await this._api.OrganizationRegistry.Create(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId },
      },
      { organization },
    )

    return Marshaler.payloadSingleResponse(response)
  }
}

export default Organizations
