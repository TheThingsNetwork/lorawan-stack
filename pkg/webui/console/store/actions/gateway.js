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

import {
  getApiKeysList,
  createGetApiKeysListActionType,
  getApiKeysListFailure,
  createGetApiKeysListFailureActionType,
  getApiKeysListSuccess,
  createGetApiKeysListSuccessActionType,
  getApiKey,
  createGetApiKeyActionType,
} from '../actions/api-keys'

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
} from '../actions/events'

export const SHARED_NAME = 'GATEWAY'

export const GET_GTW = 'GET_GATEWAY'
export const GET_GTW_SUCCESS = 'GET_GATEWAY_SUCCESS'
export const GET_GTW_FAILURE = 'GET_GATEWAY_FAILURE'
export const UPDATE_GTW = 'UPDATE_GATEWAY'
export const START_GTW_STATS = 'START_GATEWAY_STATISTICS'
export const STOP_GTW_STATS = 'STOP_GATEWAY_STATISTICS'
export const UPDATE_GTW_STATS = 'UPDATE_GATEWAY_STATISTICS'
export const UPDATE_GTW_STATS_SUCCESS = 'UPDATE_GATEWAY_STATISTICS_SUCCESS'
export const UPDATE_GTW_STATS_FAILURE = 'UPDATE_GATEWAY_STATISTICS_FAILURE'
export const UPDATE_GTW_STATS_UNAVAILABLE = 'UPDATE_GATEWAY_STATISTICS_UNAVAILABLE'
export const START_GTW_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_GTW_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_GTW_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const STOP_GTW_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_GTW_EVENTS = createClearEventsActionType(SHARED_NAME)
export const GET_GTW_API_KEYS_LIST = createGetApiKeysListActionType(SHARED_NAME)
export const GET_GTW_API_KEYS_LIST_SUCCESS = createGetApiKeysListSuccessActionType(SHARED_NAME)
export const GET_GTW_API_KEYS_LIST_FAILURE = createGetApiKeysListFailureActionType(SHARED_NAME)
export const GET_GTW_API_KEY_PAGE_DATA = createGetApiKeyActionType(SHARED_NAME)

export const getGateway = (id, meta) => (
  { type: GET_GTW, id, meta }
)

export const getGatewaySuccess = gateway => (
  { type: GET_GTW_SUCCESS, gateway }
)

export const getGatewayFailure = error => (
  { type: GET_GTW_FAILURE, error }
)

export const updateGateway = (id, patch) => (
  { type: UPDATE_GTW, id, patch }
)

export const startGatewayStatistics = (id, meta) => (
  { type: START_GTW_STATS, id, meta }
)

export const updateGatewayStatistics = id => (
  { type: UPDATE_GTW_STATS, id }
)

export const updateGatewayStatisticsSuccess = statistics => (
  { type: UPDATE_GTW_STATS_SUCCESS, statistics }
)

export const updateGatewayStatisticsFailure = error => (
  { type: UPDATE_GTW_STATS_FAILURE, error }
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

export const getGatewayApiKeysList = getApiKeysList(SHARED_NAME)

export const getGatewayApiKeysListSuccess = getApiKeysListSuccess(SHARED_NAME)

export const getGatewayApiKeysListFailure = getApiKeysListFailure(SHARED_NAME)

export const getGatewayApiKeyPageData = getApiKey(SHARED_NAME)
