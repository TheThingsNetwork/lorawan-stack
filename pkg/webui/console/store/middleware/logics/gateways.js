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

import tts from '@console/api/tts'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectGsConfig } from '@ttn-lw/lib/selectors/env'
import { getGatewayId } from '@ttn-lw/lib/selectors/id'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as gateways from '@console/store/actions/gateways'

import {
  selectGatewayById,
  selectGatewayStatisticsIsFetching,
} from '@console/store/selectors/gateways'

import createEventsConnectLogics from './events'

const createGatewayLogic = createRequestLogic({
  type: gateways.CREATE_GTW,
  process: async ({ action }) => {
    const { ownerId, gateway, isUserOwner } = action.payload

    return await tts.Gateways.create(ownerId, gateway, isUserOwner)
  },
})

const getGatewayLogic = createRequestLogic({
  type: gateways.GET_GTW,
  process: async ({ action }, dispatch) => {
    const { payload, meta } = action
    const { id = {} } = payload
    const selector = meta.selector || ''
    const gtw = await tts.Gateways.getById(id, selector)
    dispatch(gateways.startGatewayEventsStream(id))

    return gtw
  },
})

const updateGatewayLogic = createRequestLogic({
  type: gateways.UPDATE_GTW,
  process: async ({ action }) => {
    const {
      payload: { id, patch },
    } = action
    const result = await tts.Gateways.updateById(id, patch)

    return { ...patch, ...result }
  },
})

const deleteGatewayLogic = createRequestLogic({
  type: gateways.DELETE_GTW,
  process: async ({ action }) => {
    const { id } = action.payload
    const { options } = action.meta

    if (options.purge) {
      await tts.Gateways.purgeById(id)
    } else {
      await tts.Gateways.deleteById(id)
    }

    return { id }
  },
})

const restoreGatewayLogic = createRequestLogic({
  type: gateways.RESTORE_GTW,
  process: async ({ action }) => {
    const { id } = action.payload

    await tts.Gateways.restoreById(id)

    return { id }
  },
})

const getGatewaysLogic = createRequestLogic({
  type: gateways.GET_GTWS_LIST,
  latest: true,
  process: async ({ action }) => {
    const {
      params: { page, limit, query, order, deleted },
    } = action.payload
    const { selectors, options } = action.meta

    const data = options.isSearch
      ? await tts.Gateways.search(
          {
            page,
            limit,
            query,
            order,
            deleted,
          },
          selectors,
        )
      : await tts.Gateways.getAll({ page, limit, order }, selectors)

    let entities = data.gateways
    if (options.withStatus) {
      const gsConfig = selectGsConfig()
      const consoleGsAddress = getHostFromUrl(gsConfig.base_url)
      const gatewayIds = entities.map(e => e.ids)
      let gatewaysStats = null
      if (gatewayIds.length) {
        gatewaysStats = await tts.Gateways.getBatchStatistics(gatewayIds)
      }

      entities = data.gateways.map(gateway => {
        const gatewayServerAddress = getHostFromUrl(gateway.gateway_server_address)

        if (!Boolean(gatewayServerAddress)) {
          return { ...gateway, status: 'unknown' }
        }

        if (gatewayServerAddress !== consoleGsAddress) {
          return { ...gateway, status: 'other-cluster' }
        }

        if (!gatewaysStats?.entries) {
          return { ...gateway, status: 'disconnected' }
        }

        const id = getGatewayId(gateway)
        let status = 'unknown'

        if (Boolean(gatewaysStats.entries[id]) && Boolean(gatewaysStats.entries[id].connected_at)) {
          status = 'connected'
        } else if (
          !Boolean(gatewaysStats.entries[id]) ||
          Boolean(gatewaysStats.entries[id].disconnected_at)
        ) {
          status = 'disconnected'
        }

        return { ...gateway, status }
      })
    }

    return {
      entities,
      totalCount: data.totalCount,
    }
  },
})

const getGatewaysRightsLogic = createRequestLogic({
  type: gateways.GET_GTWS_RIGHTS_LIST,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await tts.Gateways.getRightsById(id)

    return result.rights.sort()
  },
})

const startGatewayStatisticsLogic = createLogic({
  type: gateways.START_GTW_STATS,
  cancelType: [gateways.STOP_GTW_STATS, gateways.UPDATE_GTW_STATS_FAILURE],
  warnTimeout: 0,
  processOptions: {
    dispatchMultiple: true,
  },
  process: async ({ cancelled$, action, getState }, dispatch, done) => {
    const { id } = action.payload
    const { timeout = 60000 } = action.meta

    const gsConfig = selectGsConfig()
    const gtw = selectGatewayById(getState(), id)

    if (!gsConfig.enabled) {
      dispatch(
        gateways.startGatewayStatisticsFailure({
          message: 'Unavailable',
        }),
      )
      done()
    }

    let gtwGsAddress
    let consoleGsAddress
    try {
      gtwGsAddress = getHostFromUrl(gtw.gateway_server_address)

      if (!Boolean(gtwGsAddress)) {
        throw new Error()
      }

      consoleGsAddress = getHostFromUrl(gsConfig.base_url)
    } catch (error) {
      dispatch(
        gateways.startGatewayStatisticsFailure({
          message: sharedMessages.statusUnknown,
        }),
      )
      done()
    }

    if (gtwGsAddress !== consoleGsAddress) {
      dispatch(
        gateways.startGatewayStatisticsFailure({
          message: sharedMessages.otherCluster,
        }),
      )
      done()
    }

    dispatch(gateways.startGatewayStatisticsSuccess())
    dispatch(gateways.updateGatewayStatistics(id))

    const interval = setInterval(() => {
      const statsRequestInProgress = selectGatewayStatisticsIsFetching(getState())
      if (!statsRequestInProgress) {
        dispatch(gateways.updateGatewayStatistics(id))
      }
    }, timeout)

    cancelled$.subscribe(() => clearInterval(interval))
  },
})

const updateGatewayStatisticsLogic = createRequestLogic({
  type: gateways.UPDATE_GTW_STATS,
  throttle: 1000,
  process: async ({ action }) => {
    const { id } = action.payload

    const stats = await tts.Gateways.getStatisticsById(id)

    return { stats }
  },
})

const getGatewayEventLocationLogic = createLogic({
  type: gateways.GET_GTW_EVENT_MESSAGE_SUCCESS,
  validate: ({ action }, allow, reject) => {
    if (action?.event?.name !== 'gs.status.receive' || !action?.event?.data?.antenna_locations) {
      reject(action)
    } else {
      allow(action)
    }
  },
  process: async ({ action }, dispatch, done) => {
    dispatch(gateways.updateGatewayLocationSuccess(action))

    done()
  },
})

export default [
  createGatewayLogic,
  getGatewayLogic,
  updateGatewayLogic,
  deleteGatewayLogic,
  restoreGatewayLogic,
  getGatewaysLogic,
  getGatewaysRightsLogic,
  startGatewayStatisticsLogic,
  updateGatewayStatisticsLogic,
  getGatewayEventLocationLogic,
  ...createEventsConnectLogics(gateways.SHARED_NAME, 'gateways', tts.Gateways.openStream),
]
