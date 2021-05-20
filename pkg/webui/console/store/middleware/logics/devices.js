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

import { STACK_COMPONENTS_MAP } from 'ttn-lw'

import api from '@console/api'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as devices from '@console/store/actions/devices'
import * as deviceTemplateFormats from '@console/store/actions/device-template-formats'

import createEventsConnectLogics from './events'

const getDeviceLogic = createRequestLogic({
  type: devices.GET_DEV,
  process: async ({ action }, dispatch) => {
    const {
      payload: { appId, deviceId },
      meta: { selector },
    } = action
    const dev = await api.device.get(appId, deviceId, selector)
    dispatch(devices.startDeviceEventsStream(dev.ids))
    return dev
  },
})

const updateDeviceLogic = createRequestLogic(
  {
    type: devices.UPDATE_DEV,
    process: async ({ action }) => {
      const {
        payload: { appId, deviceId, patch },
      } = action
      const result = await api.device.update(appId, deviceId, patch)

      return { ...patch, ...result }
    },
  },
  devices.updateDeviceSuccess,
)

const getDevicesListLogic = createRequestLogic({
  type: devices.GET_DEVICES_LIST,
  process: async ({ action }) => {
    const {
      id: appId,
      params: { page, limit, order, query },
    } = action.payload
    const { selectors, options } = action.meta

    const data = query
      ? await api.devices.search(
          appId,
          {
            page,
            limit,
            id_contains: query,
            order,
          },
          selectors,
        )
      : await api.devices.list(appId, { page, limit, order }, selectors)

    if (options.withLastSeen) {
      const macStateFetching = data.end_devices.map(async device => {
        const deviceResult = await api.device.get(appId, device.ids.device_id, 'mac_state', [
          STACK_COMPONENTS_MAP.ns,
        ])
        if ('mac_state' in deviceResult) {
          device.mac_state = deviceResult.mac_state
        }
      })

      await Promise.all(macStateFetching)
    }

    return { entities: data.end_devices, totalCount: data.totalCount }
  },
})

const getDeviceTemplateFormatsLogic = createRequestLogic({
  type: deviceTemplateFormats.GET_DEVICE_TEMPLATE_FORMATS,
  process: async () => {
    const formats = await api.deviceTemplates.listFormats()
    return formats
  },
})

export default [
  getDevicesListLogic,
  getDeviceTemplateFormatsLogic,
  getDeviceLogic,
  updateDeviceLogic,
  ...createEventsConnectLogics(devices.SHARED_NAME, 'devices', api.device.eventsSubscribe),
]
