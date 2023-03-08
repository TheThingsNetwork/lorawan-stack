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

import { APPLICATION } from '@console/constants/entities'

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
  createSetEventsFilterActionType,
  setEventsFilter,
} from './events'

export const SHARED_NAME = APPLICATION

export const CREATE_APP_BASE = 'CREATE_APP'
export const [
  { request: CREATE_APP, success: CREATE_APP_SUCCESS, failure: CREATE_APP_FAILURE },
  { request: createApp, succedd: createAppSuccess, failure: createAppFailure },
] = createRequestActions(CREATE_APP_BASE, (ownerId, app, isUserOwner) => ({
  ownerId,
  app,
  isUserOwner,
}))

export const GET_APP_BASE = 'GET_APPLICATION'
export const [
  { request: GET_APP, success: GET_APP_SUCCESS, failure: GET_APP_FAILURE },
  { request: getApplication, success: getApplicationSuccess, failure: getApplicationFailure },
] = createRequestActions(
  GET_APP_BASE,
  id => ({ id }),
  (id, selector) => ({ selector }),
)

export const ISSUE_DEV_EUI_BASE = 'ISSUE_DEV_EUI'
export const [
  { request: ISSUE_DEV_EUI, success: ISSUE_DEV_EUI_SUCCESS, failure: ISSUE_DEV_EUI_FAILURE },
  { request: issueDevEUI, success: issueDevEUISuccess, failure: issueDevEUIFailure },
] = createRequestActions(ISSUE_DEV_EUI_BASE, id => ({ id }))

export const GET_APP_DEV_EUI_COUNT_BASE = 'GET_APP_DEV_EUI_COUNT'
export const [
  {
    request: GET_APP_DEV_EUI_COUNT,
    success: GET_APP_DEV_EUI_COUNT_SUCCESS,
    failure: GET_APP_DEV_EUI_COUNT_FAILURE,
  },
  {
    request: getApplicationDevEUICount,
    success: getApplicationDevEUICountSuccess,
    failure: getApplicationDevEUICountFailure,
  },
] = createRequestActions(GET_APP_DEV_EUI_COUNT_BASE, id => ({ id }))

export const GET_APP_DEV_COUNT_BASE = 'GET_APPLICATION_DEVICE_COUNT'
export const [
  {
    request: GET_APP_DEV_COUNT,
    success: GET_APP_DEV_COUNT_SUCCESS,
    failure: GET_APP_DEV_COUNT_FAILURE,
  },
  {
    request: getApplicationDeviceCount,
    success: getApplicationDeviceCountSuccess,
    failure: getApplicationDeviceCountFailure,
  },
] = createRequestActions(GET_APP_DEV_COUNT_BASE, id => ({ id }))

export const UPDATE_APP_BASE = 'UPDATE_APPLICATION'
export const [
  { request: UPDATE_APP, success: UPDATE_APP_SUCCESS, failure: UPDATE_APP_FAILURE },
  {
    request: updateApplication,
    success: updateApplicationSuccess,
    failure: updateApplicationFailure,
  },
] = createRequestActions(UPDATE_APP_BASE, (id, patch) => ({ id, patch }))

export const DELETE_APP_BASE = createPaginationDeleteBaseActionType(SHARED_NAME)
export const [
  { request: DELETE_APP, success: DELETE_APP_SUCCESS, failure: DELETE_APP_FAILURE },
  {
    request: deleteApplication,
    success: deleteApplicationSuccess,
    failure: deleteApplicationFailure,
  },
] = createPaginationDeleteActions(SHARED_NAME, id => ({ id }))

export const RESTORE_APP_BASE = createPaginationRestoreBaseActionType(SHARED_NAME)
export const [
  { request: RESTORE_APP, success: RESTORE_APP_SUCCESS, failure: RESTORE_APP_FAILURE },
  {
    request: restoreApplication,
    success: restoreApplicationSuccess,
    failure: restoreApplicationFailure,
  },
] = createPaginationRestoreActions(SHARED_NAME, id => ({ id }))

export const GET_TOTAL_APPLICATION_COUNT_BASE = 'GET_TOTAL_APPLICATION_COUNT'
export const [
  {
    request: GET_TOTAL_APPLICATION_COUNT,
    success: GET_TOTAL_APPLICATION_COUNT_SUCCESS,
    failure: GET_TOTAL_APPLICATION_COUNT_FAILURE,
  },
  {
    request: getTotalApplicationCount,
    success: getTotalApplicationCountSuccess,
    failure: getTotalApplicationCountFailure,
  },
] = createRequestActions(GET_TOTAL_APPLICATION_COUNT_BASE)

export const GET_APPS_LIST_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  { request: GET_APPS_LIST, success: GET_APPS_LIST_SUCCESS, failure: GET_APPS_LIST_FAILURE },
  {
    request: getApplicationsList,
    success: getApplicationsSuccess,
    failure: getApplicationsFailure,
  },
] = createPaginationRequestActions(SHARED_NAME)

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

export const START_APP_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_APP_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_APP_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const PAUSE_APP_EVENT_STREAM = createPauseEventsStreamActionType(SHARED_NAME)
export const RESUME_APP_EVENT_STREAM = createResumeEventsStreamActionType(SHARED_NAME)
export const STOP_APP_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_APP_EVENTS = createClearEventsActionType(SHARED_NAME)
export const SET_APP_EVENTS_FILTER = createSetEventsFilterActionType(SHARED_NAME)
export const GET_APP_EVENT_MESSAGE_SUCCESS = createGetEventMessageSuccessActionType(SHARED_NAME)

export const startApplicationEventsStream = startEventsStream(SHARED_NAME)
export const startApplicationEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)
export const startApplicationEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)
export const pauseApplicationEventsStream = pauseEventsStream(SHARED_NAME)
export const resumeApplicationEventsStream = resumeEventsStream(SHARED_NAME)
export const stopApplicationEventsStream = stopEventsStream(SHARED_NAME)
export const clearApplicationEventsStream = clearEvents(SHARED_NAME)
export const setApplicationEventsFilter = setEventsFilter(SHARED_NAME)

export const GET_MQTT_INFO_BASE = 'GET_MQTT_INFO'
export const [
  { request: GET_MQTT_INFO, success: GET_MQTT_INFO_SUCCESS, failure: GET_MQTT_INFO_FAILURE },
  { request: getMqttInfo, success: getMqttInfoSuccess, failure: getMqttInfoFailure },
] = createRequestActions(GET_MQTT_INFO_BASE, id => ({ id }))
