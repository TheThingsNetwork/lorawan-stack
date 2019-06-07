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

import { createLogic } from 'redux-logic'

import api from '../../../api'
import * as device from '../../actions/device'
import createEventsConnectLogics from './events'

const getDeviceLogic = createLogic({
  type: [ device.GET_DEV ],
  async process ({ getState, action }, dispatch, done) {
    const { appId, deviceId, selector, options } = action
    try {
      const dev = await api.device.get(appId, deviceId, selector, options)
      dispatch(device.startDeviceEventsStream(dev.ids))
      dispatch(device.getDeviceSuccess(dev))
    } catch (e) {
      dispatch(device.getDeviceFailure(e))
    }

    done()
  },
})

export default [
  getDeviceLogic,
  ...createEventsConnectLogics(device.SHARED_NAME, 'device'),
]
