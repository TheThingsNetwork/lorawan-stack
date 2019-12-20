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

import createGetRightsListRequestActions, { createGetRightsListActionType } from './rights'
import {
  createPaginationRequestActions,
  createPaginationBaseActionType,
  createPaginationDeleteBaseActionType,
  createPaginationDeleteActions,
} from './pagination'
import { createRequestActions } from './lib'

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

export const SHARED_NAME = 'APPLICATIONS'
export const SHARED_NAME_SINGLE = 'APPLICATION'

export const GET_APP_BASE = 'GET_APPLICATION'
export const [
  { request: GET_APP, success: GET_APP_SUCCESS, failure: GET_APP_FAILURE },
  { request: getApplication, success: getApplicationSuccess, failure: getApplicationFailure },
] = createRequestActions(GET_APP_BASE, id => ({ id }), (id, selector) => ({ selector }))

export const UPDATE_APP_BASE = 'UPDATE_APPLICATION'
export const [
  { request: UPDATE_APP, success: UPDATE_APP_SUCCESS, failure: UPDATE_APP_FAILURE },
  {
    request: updateApplication,
    success: updateApplicationSuccess,
    failure: updateApplicationFailure,
  },
] = createRequestActions(UPDATE_APP_BASE, (id, patch) => ({ id, patch }))

export const DELETE_APP_BASE = createPaginationDeleteBaseActionType(SHARED_NAME_SINGLE)
export const [
  { request: DELETE_APP, success: DELETE_APP_SUCCESS, failure: DELETE_APP_FAILURE },
  {
    request: deleteApplication,
    success: deleteApplicationSuccess,
    failure: deleteApplicationFailure,
  },
] = createPaginationDeleteActions(SHARED_NAME_SINGLE, id => ({ id }))

export const GET_APPS_LIST_BASE = createPaginationBaseActionType(SHARED_NAME_SINGLE)
export const [
  { request: GET_APPS_LIST, success: GET_APPS_LIST_SUCCESS, failure: GET_APPS_LIST_FAILURE },
  {
    request: getApplicationsList,
    success: getApplicationsSuccess,
    failure: getApplicationsFailure,
  },
] = createPaginationRequestActions(SHARED_NAME_SINGLE)

export const GET_APPS_RIGHTS_LIST_BASE = createGetRightsListActionType(SHARED_NAME)
export const [
  {
    request: GET_APPS_RIGHTS_LIST,
    success: GET_APPS_RIGHTS_LIST_SUCCESS,
    failure: GET_APPS_RIGHTS_LIST_FAILURE,
  },
  {
    request: getApplicationsRightsList,
    success: getApplicationsRightsListSuccess,
    failure: getApplicationsRightsListFailure,
  },
] = createGetRightsListRequestActions(SHARED_NAME)

export const START_APP_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME_SINGLE)
export const START_APP_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(
  SHARED_NAME_SINGLE,
)
export const START_APP_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(
  SHARED_NAME_SINGLE,
)
export const STOP_APP_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME_SINGLE)
export const CLEAR_APP_EVENTS = createClearEventsActionType(SHARED_NAME_SINGLE)

export const startApplicationEventsStream = startEventsStream(SHARED_NAME_SINGLE)
export const startApplicationEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME_SINGLE)
export const startApplicationEventsStreamFailure = startEventsStreamFailure(SHARED_NAME_SINGLE)
export const stopApplicationEventsStream = stopEventsStream(SHARED_NAME_SINGLE)
export const clearApplicationEventsStream = clearEvents(SHARED_NAME_SINGLE)
