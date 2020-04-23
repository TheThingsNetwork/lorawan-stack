// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

class DownlinkQueue {
  constructor(api, { stackConfig }) {
    this._api = api
    this._stackConfig = stackConfig
  }

  async list(applicationId, deviceId) {
    const result = await this._api.DownlinkQueueList({
      routeParams: {
        'application_ids.application_id': applicationId,
        device_id: deviceId,
      },
    })

    return Marshaler.payloadListResponse('downlinks', result)
  }

  async push(applicationId, deviceId, downlinks) {
    const result = await this._api.DownlinkQueuePush(
      {
        routeParams: {
          'end_device_ids.application_ids.application_id': applicationId,
          'end_device_ids.device_id': deviceId,
        },
      },
      {
        downlinks,
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async replace(applicationId, deviceId, downlinks) {
    const result = await this._api.DownlinkQueueReplace(
      {
        routeParams: {
          'end_device_ids.application_ids.application_id': applicationId,
          'end_device_ids.device_id': deviceId,
        },
      },
      {
        downlinks,
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }
}

export default DownlinkQueue
