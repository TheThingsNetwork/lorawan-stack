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

import sharedMessages from '../../../../lib/shared-messages'
import api from '../../../api'
import * as gateways from '../../actions/gateways'
import { gsConfigSelector } from '../../../../lib/selectors/env'
import { selectSelectedGateway } from '../../selectors/gateways'
import createEventsConnectLogics from './events'
import createRequestLogic from './lib'

const getGatewayLogic = createRequestLogic({
  type: gateways.GET_GTW,
  async process ({ action }) {
    const { payload, meta } = action
    const { id = {}} = payload
    const selector = meta.selector || ''
    return api.gateway.get(id, selector)
  },
})

const updateGatewayLogic = createRequestLogic({
  type: gateways.UPDATE_GTW,
  async process ({ action }) {
    const { payload: { id, patch }} = action
    const result = await api.gateway.update(id, patch)

    return { ...patch, ...result }
  },
})

const getGatewaysLogic = createRequestLogic({
  type: gateways.GET_GTWS_LIST,
  latest: true,
  async process ({ action }) {
    const { params: { page, limit, query }} = action.payload
    const { selectors } = action.meta

    const data = query
      ? await api.gateways.search({
        page,
        limit,
        id_contains: query,
        name_contains: query,
      }, selectors)
      : await api.gateways.list({ page, limit }, selectors)

    return {
      entities: data.gateways,
      totalCount: data.totalCount,
    }
  },
})

const getGatewaysRightsLogic = createRequestLogic({
  type: gateways.GET_GTWS_RIGHTS_LIST,
  async process ({ action }, dispatch, done) {
    const { id } = action.payload
    const result = await api.rights.gateways(id)
    return result.rights.sort()
  },
})

const getGatewayCollaboratorsLogic = createRequestLogic({
  type: gateways.GET_GTW_COLLABORATORS_LIST,
  async process ({ action }) {
    const { gtwId } = action.payload
    const res = await api.gateway.collaborators.list(gtwId)
    const collaborators = res.collaborators.map(function (collaborator) {
      const { ids, ...rest } = collaborator
      const isUser = !!ids.user_ids
      const collaboratorId = isUser
        ? ids.user_ids.user_id
        : ids.organization_ids.organization_id

      return {
        id: collaboratorId,
        isUser,
        ...rest,
      }
    })
    return { id: gtwId, collaborators, totalCount: res.totalCount }
  },
})

const startGatewayStatisticsLogic = createLogic({
  type: gateways.START_GTW_STATS,
  cancelType: [
    gateways.STOP_GTW_STATS,
    gateways.UPDATE_GTW_STATS_FAILURE,
    gateways.UPDATE_GTW_STATS_UNAVAILABLE,
  ],
  warnTimeout: 0,
  validate ({ getState, action }, allow, reject) {
    const gsConfig = gsConfigSelector()
    const gtw = selectSelectedGateway(getState())

    if (!gsConfig.enabled) {
      reject(gateways.updateGatewayStatisticsUnavailable())
      return
    }

    let gtwGsAddress
    let consoleGsAddress
    try {
      const gtwAddress = gtw.gateway_server_address

      if (!Boolean(gtwAddress)) {
        throw new Error()
      }

      gtwGsAddress = gtwAddress.split(':')[0]
      consoleGsAddress = new URL(gsConfig.base_url).hostname
    } catch (error) {
      reject(gateways.updateGatewayStatisticsFailure({
        message: sharedMessages.unknown,
      }))
      return
    }

    if (gtwGsAddress !== consoleGsAddress) {
      reject(gateways.updateGatewayStatisticsFailure({
        message: sharedMessages.otherCluster,
      }))
      return
    }

    const { meta = {}} = action

    let transformed = action
    if (!meta.timeout) {
      transformed = { ...action, meta: { ...meta, timeout: 5000 }}
    }

    allow(transformed)
  },
  async process ({ cancelled$, action }, dispatch, done) {
    const { id, meta } = action

    dispatch(gateways.updateGatewayStatistics(id))

    const interval = setInterval(
      () => dispatch(gateways.updateGatewayStatistics(id)),
      meta.timeout
    )

    cancelled$.subscribe(() => clearInterval(interval))
  },
})

const updateGatewayStatisticsLogic = createRequestLogic({
  type: gateways.UPDATE_GTW_STATS,
  async process ({ action }) {
    const { id } = action.payload
    return api.gateway.stats(id)
  },
})

const getGatewayApiKeysLogic = createRequestLogic({
  type: gateways.GET_GTW_API_KEYS_LIST,
  async process ({ action }) {
    const { id: gtwId, params } = action.payload
    const res = await api.gateway.apiKeys.list(gtwId, params)
    return { ...res, id: gtwId }
  },
})

const getGatewayApiKeyLogic = createRequestLogic({
  type: gateways.GET_GTW_API_KEY,
  async process ({ action }) {
    const { id: gtwId, keyId } = action.payload
    return api.gateway.apiKeys.get(gtwId, keyId)
  },
})

export default [
  getGatewayLogic,
  updateGatewayLogic,
  getGatewaysLogic,
  getGatewaysRightsLogic,
  getGatewayCollaboratorsLogic,
  startGatewayStatisticsLogic,
  updateGatewayStatisticsLogic,
  ...createEventsConnectLogics(
    gateways.SHARED_NAME,
    'gateways',
    api.gateway.eventsSubscribe,
  ),
  getGatewayApiKeysLogic,
  getGatewayApiKeyLogic,
]
