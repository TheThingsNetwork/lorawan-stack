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
import * as devices from '../../actions/devices'
import createRequestLogic from './lib'

const getDevicesListLogic = createRequestLogic({
  type: devices.GET_DEVICES_LIST,
  async process ({ action }) {
    const { appId, params: { page, limit }} = action.payload
    const { selectors } = action.meta

    const data = await api.devices.list(appId, { page, limit }, selectors)
    return { devices: data.end_devices, totalCount: data.totalCount }
  },
})

export default [
  getDevicesListLogic,
]
