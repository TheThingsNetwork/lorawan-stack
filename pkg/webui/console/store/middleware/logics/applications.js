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

import api from '@console/api'

import { isNotFoundError, isConflictError } from '@ttn-lw/lib/errors/utils'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as applications from '@console/store/actions/applications'
import * as link from '@console/store/actions/link'

import createEventsConnectLogics from './events'

const getApplicationLogic = createRequestLogic({
  type: applications.GET_APP,
  process: async ({ action }, dispatch) => {
    const {
      payload: { id },
      meta: { selector },
    } = action
    const app = await api.application.get(id, selector)
    dispatch(applications.startApplicationEventsStream(id))
    return app
  },
})

const updateApplicationLogic = createRequestLogic({
  type: applications.UPDATE_APP,
  process: async ({ action }) => {
    const { id, patch } = action.payload

    const result = await api.application.update(id, patch)

    return { ...patch, ...result }
  },
})

const deleteApplicationLogic = createRequestLogic({
  type: applications.DELETE_APP,
  process: async ({ action }) => {
    const { id } = action.payload
    const { options } = action.meta

    if (options.purge) {
      await api.application.purge(id)
    } else {
      await api.application.delete(id)
    }

    return { id }
  },
})

const getApplicationsLogic = createRequestLogic({
  type: applications.GET_APPS_LIST,
  latest: true,
  process: async ({ action }) => {
    const {
      params: { page, limit, query, order },
    } = action.payload
    const { selectors, options } = action.meta

    const data = options.isSearch
      ? await api.applications.search(
          {
            page,
            limit,
            id_contains: query,
            order,
          },
          selectors,
        )
      : await api.applications.list({ page, limit, order }, selectors)

    return { entities: data.applications, totalCount: data.totalCount }
  },
})

const getApplicationDeviceCountLogic = createRequestLogic({
  type: applications.GET_APP_DEV_COUNT,
  process: async ({ action }) => {
    const { id: appId } = action.payload
    const data = await api.devices.list(appId, { limit: 1 })

    return { applicationDeviceCount: data.totalCount }
  },
})

const getApplicationsRightsLogic = createRequestLogic({
  type: applications.GET_APPS_RIGHTS_LIST,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await api.rights.applications(id)
    return result.rights.sort()
  },
})

const getApplicationLinkLogic = createRequestLogic({
  type: link.GET_APP_LINK,
  process: async ({ action }, dispatch, done) => {
    const {
      payload: { id },
      meta: { selector = [] } = {},
    } = action

    let linkResult
    try {
      linkResult = await api.application.link.get(id, selector)

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

export default [
  getApplicationLogic,
  getApplicationDeviceCountLogic,
  updateApplicationLogic,
  deleteApplicationLogic,
  getApplicationsLogic,
  getApplicationsRightsLogic,
  getApplicationLinkLogic,
  ...createEventsConnectLogics(
    applications.SHARED_NAME,
    'applications',
    api.application.eventsSubscribe,
  ),
]
