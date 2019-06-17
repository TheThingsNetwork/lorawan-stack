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

import createApiKeysRequestActions, { createGetApiKeysListActionType } from './api-keys'
import createApiKeyRequestActions, { createGetApiKeyActionType } from './api-key'

import {
  startEventsStream,
  createStartEventsStreamActionType,
  startEventsStreamSuccess,
  createStartEventsStreamSuccessActionType,
  startEventsStreamFailure,
  createStartEventsStreamFailureActionType,
  stopEventsStream,
  createStopEventsStreamActionType,
  clearEvents,
  createClearEventsActionType,
} from './events'

import { createRequestActions } from './lib'

export const SHARED_NAME = 'GATEWAY'

export const GET_GTW_BASE = 'GET_GATEWAY'
export const [{
  request: GET_GTW,
  success: GET_GTW_SUCCESS,
  failure: GET_GTW_FAILURE,
}, {
  request: getGateway,
  success: getGatewaySuccess,
  failure: getGatewayFailure,
}] = createRequestActions(GET_GTW_BASE,
  id => ({ id }),
  (id, selector) => ({ selector })
)

export const UPDATE_GTW_STATS_BASE = 'UPDATE_GATEWAY_STATISTICS'
export const [{
  request: UPDATE_GTW_STATS,
  success: UPDATE_GTW_STATS_SUCCESS,
  failure: UPDATE_GTW_STATS_FAILURE,
}, {
  request: updateGatewayStatistics,
  success: updateGatewayStatisticsSuccess,
  failure: updateGatewayStatisticsFailure,
}] = createRequestActions(UPDATE_GTW_STATS_BASE, id => ({ id }))

export const GET_GTW_API_KEY_BASE = createGetApiKeyActionType(SHARED_NAME)
export const [{
  request: GET_GTW_API_KEY,
  success: GET_GTW_API_KEY_SUCCESS,
  failure: GET_GTW_API_KEY_FAILURE,
}, {
  request: getGatewayApiKey,
  success: getGatewayApiKeySuccess,
  failure: getGatewayApiKeyFailure,
}] = createApiKeyRequestActions(SHARED_NAME)

export const GET_GTW_API_KEYS_LIST_BASE = createGetApiKeysListActionType(SHARED_NAME)
export const [{
  request: GET_GTW_API_KEYS_LIST,
  success: GET_GTW_API_KEYS_LIST_SUCCESS,
  failure: GET_GTW_API_KEYS_LIST_FAILURE,
}, {
  request: getGatewayApiKeysList,
  success: getGatewayApiKeysListSuccess,
  failure: getGatewayApiKeysListFailure,
}] = createApiKeysRequestActions(SHARED_NAME)

export const UPDATE_GTW = 'UPDATE_GATEWAY'
export const START_GTW_STATS = 'START_GATEWAY_STATISTICS'
export const STOP_GTW_STATS = 'STOP_GATEWAY_STATISTICS'
export const UPDATE_GTW_STATS_UNAVAILABLE = 'UPDATE_GATEWAY_STATISTICS_UNAVAILABLE'
export const START_GTW_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_GTW_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_GTW_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const STOP_GTW_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_GTW_EVENTS = createClearEventsActionType(SHARED_NAME)

export const updateGateway = (id, patch) => (
  { type: UPDATE_GTW, id, patch }
)

export const startGatewayStatistics = (id, meta) => (
  { type: START_GTW_STATS, id, meta }
)

export const updateGatewayStatisticsUnavailable = () => (
  { type: UPDATE_GTW_STATS_UNAVAILABLE }
)

export const stopGatewayStatistics = () => (
  { type: STOP_GTW_STATS }
)

export const startGatewayEventsStream = startEventsStream(SHARED_NAME)

export const startGatewayEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)

export const startGatewayEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)

export const stopGatewayEventsStream = stopEventsStream(SHARED_NAME)

export const clearGatewayEventsStream = clearEvents(SHARED_NAME)
