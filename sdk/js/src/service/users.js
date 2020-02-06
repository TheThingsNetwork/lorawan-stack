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
import { createDefaultsEmitterFromFieldMask } from '../util/create-defaults-emitter'

const userDefaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
  // Ensure to set STATE_REQUESTED if needed, which gets stripped as null
  // value from the backend response
  if (fmKey === 'state' && !Boolean(value)) {
    return 'STATE_REQUESTED'
  }
})

class Users {
  constructor(registry) {
    this._api = registry
  }

  async getAll(params, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.UserRegistry.List(undefined, {
      ...params,
      ...fieldMask,
    })

    return Marshaler.payloadListResponse('users', response, user =>
      userDefaultsEmitter(user, fieldMask.field_mask.paths),
    )
  }

  async search(params, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.EntityRegistrySearch.SearchUsers(undefined, {
      ...params,
      ...fieldMask,
    })

    return Marshaler.payloadListResponse('users', response, user =>
      userDefaultsEmitter(user, fieldMask.field_mask.paths),
    )
  }

  async getById(id, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.UserRegistry.Get(
      {
        routeParams: { 'user_ids.user_id': id },
      },
      fieldMask,
    )

    return Marshaler.payloadSingleResponse(response, user =>
      userDefaultsEmitter(user, fieldMask.field_mask.paths),
    )
  }

  async deleteById(id) {
    const response = await this._api.UserRegistry.Delete({
      routeParams: { user_id: id },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async updateById(id, patch, mask = Marshaler.fieldMaskFromPatch(patch)) {
    const response = await this._api.UserRegistry.Update(
      {
        routeParams: {
          'user.ids.user_id': id,
        },
      },
      {
        user: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )
    return Marshaler.unwrapUser(response)
  }
}

export default Users
