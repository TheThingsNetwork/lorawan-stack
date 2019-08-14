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

import api from '../../../api'
import * as device from '../../actions/device'
import createEventsConnectLogics from './events'
import createRequestLogic from './lib'

const getDeviceLogic = createRequestLogic({
  type: device.GET_DEV,
  async process({ action }, dispatch) {
    const {
      payload: { appId, deviceId },
      meta: { selector, options },
    } = action
    const dev = await api.device.get(appId, deviceId, selector, options)
    dispatch(device.startDeviceEventsStream(dev.ids))
    return dev
  },
})

const updateDeviceLogic = createRequestLogic(
  {
    type: device.UPDATE_DEV,
    async process({ action }) {
      const {
        payload: { appId, deviceId, patch },
      } = action
      const result = await api.device.update(appId, deviceId, patch)

      return { ...patch, ...result }
    },
  },
  device.updateDeviceSuccess,
)

export default [
  getDeviceLogic,
  updateDeviceLogic,
  ...createEventsConnectLogics(device.SHARED_NAME, 'devices', api.device.eventsSubscribe),
]
