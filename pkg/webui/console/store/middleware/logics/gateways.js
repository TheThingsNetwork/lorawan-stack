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
import { selectGsConfig } from '../../../../lib/selectors/env'
import { selectGatewayById } from '../../selectors/gateways'
import createEventsConnectLogics from './events'
import createRequestLogic from './lib'

const getGatewayLogic = createRequestLogic({
  type: gateways.GET_GTW,
  async process ({ action }, dispatch) {
    const { payload, meta } = action
    const { id = {}} = payload
    const selector = meta.selector || ''
    const gtw = await api.gateway.get(id, selector)
    dispatch(gateways.startGatewayEventsStream(id))
    return gtw
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

const deleteGatewayLogic = createRequestLogic({
  type: gateways.DELETE_GTW,
  async process ({ action }) {
    const { id } = action.payload

    await api.gateway.delete(id)

    return { id }
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

const getGatewayCollaboratorLogic = createRequestLogic({
  type: gateways.GET_GTW_COLLABORATOR,
  async process ({ action }) {
    const { id: gtwId, collaboratorId, isUser } = action.payload

    const collaborator = isUser
      ? await api.gateway.collaborators.getUser(gtwId, collaboratorId)
      : await api.gateway.collaborators.getOrganization(gtwId, collaboratorId)

    const { ids, ...rest } = collaborator

    return {
      id: collaboratorId,
      isUser,
      ...rest,
    }
  },
})

const getGatewayCollaboratorsLogic = createRequestLogic({
  type: gateways.GET_GTW_COLLABORATORS_LIST,
  async process ({ action }) {
    const { id, params } = action.payload
    const res = await api.gateway.collaborators.list(id, params)
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
    return { id, collaborators, totalCount: res.totalCount }
  },
})

const startGatewayStatisticsLogic = createLogic({
  type: gateways.START_GTW_STATS,
  cancelType: [
    gateways.STOP_GTW_STATS,
    gateways.UPDATE_GTW_STATS_FAILURE,
  ],
  warnTimeout: 0,
  processOptions: {
    dispatchMultiple: true,
  },
  async process ({ cancelled$, action, getState }, dispatch, done) {
    const { id } = action.payload
    const { timeout = 5000 } = action.meta

    const gsConfig = selectGsConfig()
    const gtw = selectGatewayById(getState(), id)

    if (!gsConfig.enabled) {
      dispatch(gateways.startGatewayStatisticsFailure({
        message: 'Unavailable',
      }))
      done()
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
      dispatch(gateways.startGatewayStatisticsFailure({
        message: sharedMessages.unknown,
      }))
      done()
    }

    if (gtwGsAddress !== consoleGsAddress) {
      dispatch(gateways.startGatewayStatisticsFailure({
        message: sharedMessages.otherCluster,
      }))
      done()
    }

    dispatch(gateways.startGatewayStatisticsSuccess())
    dispatch(gateways.updateGatewayStatistics(id))

    const interval = setInterval(
      () => dispatch(gateways.updateGatewayStatistics(id)),
      timeout
    )

    cancelled$.subscribe(() => clearInterval(interval))
  },
})

const updateGatewayStatisticsLogic = createRequestLogic({
  type: gateways.UPDATE_GTW_STATS,
  async process ({ action }) {
    const { id } = action.payload

    const stats = await api.gateway.stats(id)

    return { stats }
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
  deleteGatewayLogic,
  getGatewaysLogic,
  getGatewaysRightsLogic,
  getGatewayCollaboratorLogic,
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
