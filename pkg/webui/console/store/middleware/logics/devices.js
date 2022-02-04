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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import { mayReadApplicationDeviceKeys, checkFromState } from '@console/lib/feature-checks'

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
    const dev = await tts.Applications.Devices.getById(appId, deviceId, selector)
    dispatch(devices.startDeviceEventsStream(dev.ids))
    return { ...dev }
  },
})

const updateDeviceLogic = createRequestLogic(
  {
    type: devices.UPDATE_DEV,
    process: async ({ action }) => {
      const {
        payload: { appId, deviceId, patch },
      } = action
      const result = await tts.Applications.Devices.updateById(appId, deviceId, patch)

      return { ...patch, ...result }
    },
  },
  devices.updateDeviceSuccess,
)

const getDevicesListLogic = createRequestLogic({
  type: devices.GET_DEVICES_LIST,
  process: async ({ action, getState }) => {
    const {
      id: appId,
      params: { page, limit, order, query },
    } = action.payload
    const { selectors, options } = action.meta

    const data = query
      ? await tts.Applications.Devices.search(
          appId,
          {
            page,
            limit,
            id_contains: query,
            order,
          },
          selectors,
        )
      : await tts.Applications.Devices.getAll(appId, { page, limit, order }, selectors)

    if (options.withLastSeen) {
      const mayReadKeys = checkFromState(mayReadApplicationDeviceKeys, getState())
      const selector = ['mac_state.recent_uplinks', 'pending_mac_state.recent_uplinks']
      if (mayReadKeys) {
        selector.push('session.started_at', 'pending_session')
      }
      const activityFetching = data.end_devices.map(async device => {
        const deviceResult = await tts.Applications.Devices.getById(
          appId,
          device.ids.device_id,
          selector,
          [STACK_COMPONENTS_MAP.ns],
        )

        // Merge activity-relevant fields into fetched device.
        if ('mac_state' in deviceResult) {
          device.mac_state = deviceResult.mac_state
        } else if ('pending_mac_state' in deviceResult) {
          device.pending_mac_state = deviceResult.pending_mac_state
        }
        if (mayReadKeys) {
          if ('session' in deviceResult) {
            device.session = deviceResult.session
          } else if ('pending_session' in deviceResult) {
            device.pending_session = deviceResult.pendingSession
          }
        }
      })

      await Promise.all(activityFetching)
    }

    return { entities: data.end_devices, totalCount: data.totalCount }
  },
})

const resetDeviceLogic = createRequestLogic({
  type: devices.RESET_DEV,
  process: async ({ action }) => {
    const { appId, deviceId } = action.payload

    return tts.Applications.Devices.resetById(appId, deviceId)
  },
})

const getDeviceTemplateFormatsLogic = createRequestLogic({
  type: deviceTemplateFormats.GET_DEVICE_TEMPLATE_FORMATS,
  process: async () => {
    const formats = await tts.Applications.Devices.listTemplateFormats()
    return formats
  },
})

export default [
  getDevicesListLogic,
  getDeviceTemplateFormatsLogic,
  getDeviceLogic,
  resetDeviceLogic,
  updateDeviceLogic,
  ...createEventsConnectLogics(devices.SHARED_NAME, 'devices', tts.Applications.Devices.openStream),
]
