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

import tts from '@console/api/tts'

import { isNotFoundError, isConflictError } from '@ttn-lw/lib/errors/utils'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as applications from '@console/store/actions/applications'
import * as link from '@console/store/actions/link'

import createEventsConnectLogics from './events'

const createApplicationLogic = createRequestLogic({
  type: applications.CREATE_APP,
  process: async ({ action }) => {
    const { ownerId, app, isAdmin } = action.payload
    const newApp = await tts.Applications.create(ownerId, app, isAdmin)
    return newApp
  },
})

const getApplicationLogic = createRequestLogic({
  type: applications.GET_APP,
  process: async ({ action }, dispatch) => {
    const {
      payload: { id },
      meta: { selector },
    } = action
    const app = await tts.Applications.getById(id, selector)
    dispatch(applications.startApplicationEventsStream(id))
    return app
  },
})

const issueDevEUILogic = createRequestLogic({
  type: applications.ISSUE_DEV_EUI,
  process: async ({ action }) => {
    const { id } = action.payload
    return await tts.Applications.issueDevEUI(id)
  },
})

const getApplicationDevEUICountLogic = createRequestLogic({
  type: applications.GET_APP_DEV_EUI_COUNT,
  process: async ({ action }) => {
    const {
      payload: { id },
    } = action
    const result = await tts.Applications.getById(id, 'dev_eui_counter')
    return { id, dev_eui_counter: result.dev_eui_counter }
  },
})

const updateApplicationLogic = createRequestLogic({
  type: applications.UPDATE_APP,
  process: async ({ action }) => {
    const { id, patch } = action.payload

    const result = await tts.Applications.updateById(id, patch)

    return { ...patch, ...result }
  },
})

const deleteApplicationLogic = createRequestLogic({
  type: applications.DELETE_APP,
  process: async ({ action }) => {
    const { id } = action.payload
    const { options } = action.meta

    if (options.purge) {
      await tts.Applications.purgeById(id)
    } else {
      await tts.Applications.deleteById(id)
    }

    return { id }
  },
})

const restoreApplicationLogic = createRequestLogic({
  type: applications.RESTORE_APP,
  process: async ({ action }) => {
    const { id } = action.payload

    await tts.Applications.restoreById(id)

    return { id }
  },
})

const getApplicationsLogic = createRequestLogic({
  type: applications.GET_APPS_LIST,
  latest: true,
  process: async ({ action }, dispatch) => {
    const {
      params: { page, limit, query, order, deleted },
    } = action.payload
    const { selectors, options } = action.meta

    const data = options.isSearch
      ? await tts.Applications.search(
          {
            page,
            limit,
            query,
            order,
            deleted,
          },
          selectors,
        )
      : await tts.Applications.getAll({ page, limit, order }, selectors)

    if (options.withDeviceCount) {
      for (const application of data.applications) {
        dispatch(applications.getApplicationDeviceCount(application.ids.application_id))
      }
    }

    return { entities: data.applications, totalCount: data.totalCount }
  },
})

const getApplicationDeviceCountLogic = createRequestLogic({
  type: applications.GET_APP_DEV_COUNT,
  process: async ({ action }) => {
    const { id: appId } = action.payload
    const data = await tts.Applications.Devices.getAll(appId, { limit: 1 })

    return { id: appId, applicationDeviceCount: data.totalCount }
  },
})

const getApplicationsRightsLogic = createRequestLogic({
  type: applications.GET_APPS_RIGHTS_LIST,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await tts.Applications.getRightsById(id)
    return result.rights.sort()
  },
})

const getApplicationLinkLogic = createRequestLogic({
  type: link.GET_APP_LINK,
  process: async ({ action }) => {
    const {
      payload: { id },
      meta: { selector = [] } = {},
    } = action

    let linkResult
    try {
      linkResult = await tts.Applications.Link.get(id, selector)

      return { link: linkResult }
    } catch (error) {
      // Ignore 404 error. It means that the application is not linked, but the response can
      // still hold link data that we have to display to the user.
      if (isNotFoundError(error) && typeof linkResult !== 'undefined') {
        return { link: linkResult }
      }

      // Ignore 409 error. It means that the application link cannot be established, but
      // the response can still hold link data that we have to displat to the user.
      if (isConflictError(error) && typeof linkResult !== 'undefined') {
        return { link: linkResult }
      }

      throw error
    }
  },
})

const updateApplicationLinkLogic = createRequestLogic(
  {
    type: link.UPDATE_APP_LINK,
    process: async ({ action }) => {
      const { id, link } = action.payload

      const updatedLink = await tts.Applications.Link.set(id, link)

      return { ...link, ...updatedLink }
    },
  },
  link.updateApplicationLinkSuccess,
)

const getMqttConnectionInfoLogic = createRequestLogic({
  type: applications.GET_MQTT_INFO,
  process: async ({ action }) => {
    const { id } = action.payload

    const mqttInfo = await tts.Applications.getMqttConnectionInfo(id)

    return mqttInfo
  },
})

export default [
  createApplicationLogic,
  getApplicationLogic,
  getApplicationDeviceCountLogic,
  updateApplicationLogic,
  deleteApplicationLogic,
  restoreApplicationLogic,
  getApplicationsLogic,
  getApplicationsRightsLogic,
  getApplicationLinkLogic,
  updateApplicationLinkLogic,
  issueDevEUILogic,
  getApplicationDevEUICountLogic,
  getMqttConnectionInfoLogic,
  ...createEventsConnectLogics(
    applications.SHARED_NAME,
    'applications',
    tts.Applications.openStream,
  ),
]
