// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

class As {
  constructor(service) {
    this._api = service
  }

  async encodeDownlink(appId, deviceId, data) {
    const result = await this._api.EncodeDownlink(
      {
        routeParams: {
          'end_device_ids.application_ids.application_id': appId,
          'end_device_ids.device_id': deviceId,
        },
      },
      data,
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async decodeDownlink(appId, deviceId, data) {
    const result = await this._api.DecodeDownlink(
      {
        routeParams: {
          'end_device_ids.application_ids.application_id': appId,
          'end_device_ids.device_id': deviceId,
        },
      },
      data,
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async decodeUplink(appId, deviceId, data) {
    const result = await this._api.DecodeUplink(
      {
        routeParams: {
          'end_device_ids.application_ids.application_id': appId,
          'end_device_ids.device_id': deviceId,
        },
      },
      data,
    )

    return Marshaler.payloadSingleResponse(result)
  }
}

export default As
