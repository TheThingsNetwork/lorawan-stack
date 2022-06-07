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
import { defineMessage } from 'react-intl'

import tts from '@console/api/tts'

import toast from '@ttn-lw/components/toast'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as devices from '@console/store/actions/devices'
import * as deviceTemplateFormats from '@console/store/actions/device-template-formats'

import { selectDeviceByIds } from '@console/store/selectors/devices'

import createEventsConnectLogics from './events'

const m = defineMessage({
  joinSuccess: 'The device has successfully joined the network',
})

const getDeviceLogic = createRequestLogic({
  type: devices.GET_DEV,
  process: async ({ action }, dispatch) => {
    const {
      payload: { appId, deviceId },
      meta: { selector },
    } = action
    const dev = await tts.Applications.Devices.getById(appId, deviceId, selector)
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
      const result = await tts.Applications.Devices.updateById(appId, deviceId, patch)

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
    const { selectors } = action.meta

    const data = query
      ? await tts.Applications.Devices.search(
          appId,
          {
            page,
            limit,
            query,
            order,
          },
          selectors,
        )
      : await tts.Applications.Devices.getAll(appId, { page, limit, order }, selectors)

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
  process: async () => await tts.Applications.Devices.listTemplateFormats(),
})

const convertTemplateLogic = createRequestLogic({
  type: deviceTemplateFormats.CONVERT_TEMPLATE,
  process: async ({ action }) => {
    const { format_id, data } = action.payload

    return await tts.Applications.Devices.convertTemplate(format_id, data)
  },
})

const getDeviceSessionLogic = createLogic({
  type: devices.GET_DEVICE_EVENT_MESSAGE_SUCCESS,
  process: async ({ action, getState }, dispatch, done) => {
    const { event, id } = action
    const appId = id.application_ids.application_id
    const sessionSelector = ['pending_session', 'session']

    const device = selectDeviceByIds(getState(), appId, id.device_id)

    if (event.name === 'as.up.join.forward') {
      const dev = await tts.Applications.Devices.getById(appId, id.device_id, sessionSelector)
      dispatch(devices.getDeviceSuccess(dev))
    }

    if (
      (event.name === 'ns.up.data.process' || event.name === 'as.up.data.process') &&
      device.pending_session !== null
    ) {
      const dev = await tts.Applications.Devices.getById(appId, id.device_id, sessionSelector)
      dispatch(devices.getDeviceSuccess(dev))
      if (!('pending_session' in dev) && 'session' in dev) {
        toast({
          title: id.device_id,
          message: m.joinSuccess,
          type: toast.types.INFO,
        })
      }
    }

    done()
  },
})

export default [
  getDevicesListLogic,
  getDeviceTemplateFormatsLogic,
  convertTemplateLogic,
  getDeviceLogic,
  resetDeviceLogic,
  updateDeviceLogic,
  getDeviceSessionLogic,
  ...createEventsConnectLogics(devices.SHARED_NAME, 'devices', tts.Applications.Devices.openStream),
]
