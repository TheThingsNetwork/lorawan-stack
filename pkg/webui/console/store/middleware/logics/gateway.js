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
import * as gateway from '../../actions/gateway'
import { gsConfigSelector } from '../../../../lib/selectors/env'
import { selectSelectedGateway } from '../../selectors/gateway'
import createEventsConnectLogics from './events'
import createRequestLogic from './lib'

const getGatewayLogic = createRequestLogic({
  type: gateway.GET_GTW,
  async process ({ action }) {
    const { payload, meta } = action
    const { id = {}} = payload
    const selector = meta.selector || ''
    return api.gateway.get(id, selector)
  },
})

const startGatewayStatisticsLogic = createLogic({
  type: gateway.START_GTW_STATS,
  cancelType: [
    gateway.STOP_GTW_STATS,
    gateway.UPDATE_GTW_STATS_FAILURE,
    gateway.UPDATE_GTW_STATS_UNAVAILABLE,
  ],
  warnTimeout: 0,
  validate ({ getState, action }, allow, reject) {
    const gsConfig = gsConfigSelector()
    const gtw = selectSelectedGateway(getState())

    if (!gsConfig.enabled) {
      reject(gateway.updateGatewayStatisticsUnavailable())
      return
    }

    const gtwGsAddress = gtw.gateway_server_address
    const consoleGsAddress = new URL(gsConfig.base_url).host

    if (!Boolean(gtwGsAddress)) {
      reject(gateway.updateGatewayStatisticsFailure({
        message: sharedMessages.unknown,
      }))
      return
    }

    if (gtwGsAddress !== consoleGsAddress) {
      reject(gateway.updateGatewayStatisticsFailure({
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

    dispatch(gateway.updateGatewayStatistics(id))

    const interval = setInterval(
      () => dispatch(gateway.updateGatewayStatistics(id)),
      meta.timeout
    )

    cancelled$.subscribe(() => clearInterval(interval))
  },
})

const updateGatewayStatisticsLogic = createRequestLogic({
  type: gateway.UPDATE_GTW_STATS,
  async process ({ action }) {
    const { id } = action.payload
    return api.gateway.stats(id)
  },
})

const getGatewayApiKeysLogic = createRequestLogic({
  type: gateway.GET_GTW_API_KEYS_LIST,
  async process ({ action }) {
    const { id: gtwId, params } = action.payload
    const res = await api.gateway.apiKeys.list(gtwId, params)
    return { ...res, id: gtwId }
  },
})

const getGatewayApiKeyLogic = createRequestLogic({
  type: gateway.GET_GTW_API_KEY,
  async process ({ action }) {
    const { id: gtwId, keyId } = action.payload
    return api.gateway.apiKeys.get(gtwId, keyId)
  },
})

export default [
  getGatewayLogic,
  startGatewayStatisticsLogic,
  updateGatewayStatisticsLogic,
  ...createEventsConnectLogics(
    gateway.SHARED_NAME,
    'gateways',
    api.gateway.eventsSubscribe,
  ),
  getGatewayApiKeysLogic,
  getGatewayApiKeyLogic,
]
