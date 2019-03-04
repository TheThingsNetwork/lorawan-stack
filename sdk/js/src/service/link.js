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

class Link {
  constructor (registry) {
    this._api = registry

    this.get = this.get.bind(this)
    this.set = this.set.bind(this)
    this.delete = this.delete.bind(this)
    this.getStats = this.getStats.bind(this)
  }

  async get (appId, fieldMask) {
    const result = await this._api.GetLink({
      route: { 'application_ids.application_id': appId },
      query: Marshaler.queryFieldMask(fieldMask),
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async set (appId, data, mask = Marshaler.fieldMaskFromPatch(data)) {
    const result = await this._api.SetLink({
      route: { 'application_ids.application_id': appId },
    },
    {
      link: data,
      field_mask: Marshaler.fieldMask(mask),
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async delete (appId) {
    const result = await this._api.DeleteLink({
      route: { application_id: appId },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getStats (appId) {
    const result = await this._api.GetLinkStats({
      route: { application_id: appId },
    })

    return Marshaler.payloadSingleResponse(result)
  }
}

export default Link
