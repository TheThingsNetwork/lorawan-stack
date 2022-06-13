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

import { GATEWAY } from '@console/constants/entities'

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationRequestActions,
  createPaginationBaseActionType,
  createPaginationDeleteBaseActionType,
  createPaginationDeleteActions,
  createPaginationRestoreBaseActionType,
  createPaginationRestoreActions,
} from '@ttn-lw/lib/store/actions/pagination'

import createGetRightsListRequestActions, { createGetRightsListActionType } from './rights'
import {
  startEventsStream,
  createStartEventsStreamActionType,
  startEventsStreamSuccess,
  createStartEventsStreamSuccessActionType,
  startEventsStreamFailure,
  createStartEventsStreamFailureActionType,
  pauseEventsStream,
  createPauseEventsStreamActionType,
  resumeEventsStream,
  createResumeEventsStreamActionType,
  stopEventsStream,
  createStopEventsStreamActionType,
  clearEvents,
  createClearEventsActionType,
  createGetEventMessageSuccessActionType,
  getEventMessageSuccess,
  createSetEventsFilterActionType,
  setEventsFilter,
} from './events'

export const SHARED_NAME = GATEWAY

export const CREATE_GTW_BASE = 'CREATE_GTW'
export const [
  { request: CREATE_GTW, success: CREATE_GTW_SUCCESS, failure: CREATE_GTW_FAILURE },
  { request: createGateway, success: createGatewaySuccess, failure: createGatewayFailure },
] = createRequestActions(CREATE_GTW_BASE, (ownerId, gateway, isUserOwner) => ({
  ownerId,
  gateway,
  isUserOwner,
}))

export const GET_GTW_BASE = 'GET_GATEWAY'
export const [
  { request: GET_GTW, success: GET_GTW_SUCCESS, failure: GET_GTW_FAILURE },
  { request: getGateway, success: getGatewaySuccess, failure: getGatewayFailure },
] = createRequestActions(
  GET_GTW_BASE,
  id => ({ id }),
  (id, selector) => ({ selector }),
)

export const UPDATE_GTW_BASE = 'UPDATE_GATEWAY'
export const [
  { request: UPDATE_GTW, success: UPDATE_GTW_SUCCESS, failure: UPDATE_GTW_FAILURE },
  { request: updateGateway, success: updateGatewaySuccess, failure: updateGatewayFailure },
] = createRequestActions(
  UPDATE_GTW_BASE,
  (id, patch) => ({ id, patch }),
  (id, patch, selector) => ({ selector }),
)

export const DELETE_GTW_BASE = createPaginationDeleteBaseActionType(SHARED_NAME)
export const [
  { request: DELETE_GTW, success: DELETE_GTW_SUCCESS, failure: DELETE_GTW_FAILURE },
  { request: deleteGateway, success: deleteGatewaySuccess, failure: deleteGatewayFailure },
] = createPaginationDeleteActions(SHARED_NAME)

export const RESTORE_GTW_BASE = createPaginationRestoreBaseActionType(SHARED_NAME)
export const [
  { request: RESTORE_GTW, success: RESTORE_GTW_SUCCESS, failure: RESTORE_GTW_FAILURE },
  { request: restoreGateway, success: restoreGatewaySuccess, failure: restoreGatewayFailure },
] = createPaginationRestoreActions(SHARED_NAME, id => ({ id }))

export const GET_GTWS_LIST_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  { request: GET_GTWS_LIST, success: GET_GTWS_LIST_SUCCESS, failure: GET_GTWS_LIST_FAILURE },
  { request: getGatewaysList, success: getGatewaysListSuccess, failure: getGatewaysListFailure },
] = createPaginationRequestActions(SHARED_NAME)

export const GET_GTWS_RIGHTS_LIST_BASE = createGetRightsListActionType(SHARED_NAME)
export const [
  {
    request: GET_GTWS_RIGHTS_LIST,
    success: GET_GTWS_RIGHTS_LIST_SUCCESS,
    failure: GET_GTWS_RIGHTS_LIST_FAILURE,
  },
  {
    request: getGatewaysRightsList,
    success: getGatewaysRightsListSuccess,
    failure: getGatewaysRightsListFailure,
  },
] = createGetRightsListRequestActions(SHARED_NAME)

export const UPDATE_GTW_STATS_BASE = 'UPDATE_GATEWAY_STATISTICS'
export const [
  {
    request: UPDATE_GTW_STATS,
    success: UPDATE_GTW_STATS_SUCCESS,
    failure: UPDATE_GTW_STATS_FAILURE,
  },
  {
    request: updateGatewayStatistics,
    success: updateGatewayStatisticsSuccess,
    failure: updateGatewayStatisticsFailure,
  },
] = createRequestActions(UPDATE_GTW_STATS_BASE, id => ({ id }))

export const START_GTW_STATS_BASE = 'START_GATEWAY_STATISTICS'
export const [
  { request: START_GTW_STATS, success: START_GTW_STATS_SUCCESS, failure: START_GTW_STATS_FAILURE },
  {
    request: startGatewayStatistics,
    success: startGatewayStatisticsSuccess,
    failure: startGatewayStatisticsFailure,
  },
] = createRequestActions(
  START_GTW_STATS_BASE,
  id => ({ id }),
  (id, timeout) => ({ timeout }),
)

export const STOP_GTW_STATS = 'STOP_GATEWAY_STATISTICS'
export const stopGatewayStatistics = () => ({ type: STOP_GTW_STATS })

export const START_GTW_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_GTW_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_GTW_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const PAUSE_GTW_EVENT_STREAM = createPauseEventsStreamActionType(SHARED_NAME)
export const RESUME_GTW_EVENT_STREAM = createResumeEventsStreamActionType(SHARED_NAME)
export const STOP_GTW_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_GTW_EVENTS = createClearEventsActionType(SHARED_NAME)
export const GET_GTW_EVENT_MESSAGE_SUCCESS = createGetEventMessageSuccessActionType(SHARED_NAME)
export const SET_GTW_EVENTS_FILTER = createSetEventsFilterActionType(SHARED_NAME)

export const startGatewayEventsStream = startEventsStream(SHARED_NAME)
export const startGatewayEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)
export const startGatewayEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)
export const pauseGatewayEventsStream = pauseEventsStream(SHARED_NAME)
export const resumeGatewayEventsStream = resumeEventsStream(SHARED_NAME)
export const stopGatewayEventsStream = stopEventsStream(SHARED_NAME)
export const clearGatewayEventsStream = clearEvents(SHARED_NAME)
export const getGatewayEventMessageSuccess = getEventMessageSuccess(SHARED_NAME)
export const setGatewayEventsFilter = setEventsFilter(SHARED_NAME)
