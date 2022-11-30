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

import { END_DEVICE } from '@console/constants/entities'

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationByIdRequestActions,
  createPaginationBaseActionType,
} from '@ttn-lw/lib/store/actions/pagination'

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
  createSetEventsFilterActionType,
  createGetEventMessageSuccessActionType,
  setEventsFilter,
} from './events'

export const SHARED_NAME = END_DEVICE

export const CREATE_DEVICE_BASE = 'CREATE_DEVICE'
export const [
  { request: CREATE_DEVICE, success: CREATE_DEVICE_SUCCESS, failure: CREATE_DEVICE_FAILURE },
  { request: createDevice, success: createDeviceSuccess, failure: createDeviceFailure },
] = createRequestActions(CREATE_DEVICE_BASE, (appId, device) => ({ appId, device }))

export const GET_DEV_BASE = 'GET_END_DEVICE'
export const [
  { request: GET_DEV, success: GET_DEV_SUCCESS, failure: GET_DEV_FAILURE },
  { request: getDevice, success: getDeviceSuccess, failure: getDeviceFailure },
] = createRequestActions(
  GET_DEV_BASE,
  (appId, deviceId) => ({ appId, deviceId }),
  (appId, deviceId, selector, options) => ({ selector, options }),
)

export const UPDATE_DEV_BASE = 'UPDATE_END_DEVICE'
export const [
  { request: UPDATE_DEV, success: UPDATE_DEV_SUCCESS, failure: UPDATE_DEV_FAILURE },
  { request: updateDevice, success: updateDeviceSuccess, failure: updateDeviceFailure },
] = createRequestActions(
  UPDATE_DEV_BASE,
  (appId, deviceId, patch) => ({ appId, deviceId, patch }),
  (appId, deviceId, patch, selector) => ({ selector }),
)

export const GET_DEVICES_LIST_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  {
    request: GET_DEVICES_LIST,
    success: GET_DEVICES_LIST_SUCCESS,
    failure: GET_DEVICES_LIST_FAILURE,
  },
  { request: getDevicesList, success: getDevicesListSuccess, failure: getDevicesListFailure },
] = createPaginationByIdRequestActions(SHARED_NAME)

export const RESET_DEV_BASE = 'RESET_END_DEVICE'
export const [
  { request: RESET_DEV, success: RESET_DEV_SUCCESS, failure: RESET_DEV_FAILURE },
  { request: resetDevice, success: resetDeviceSuccess, failure: resetDeviceFailure },
] = createRequestActions(RESET_DEV_BASE, (appId, deviceId) => ({ appId, deviceId }))

export const RESET_USED_DEV_NONCES_BASE = 'RESET_USED_DEV_NONCES'
export const [
  {
    request: RESET_USED_DEV_NONCES,
    success: RESET_USED_DEV_NONCES_SUCCESS,
    failure: RESET_USED_DEV_NONCES_FAILURE,
  },
  {
    request: resetUsedDevNonces,
    success: resetUsedDevNoncesSuccess,
    failure: resetUsedDevNoncesFailure,
  },
] = createRequestActions(RESET_USED_DEV_NONCES_BASE, (appId, deviceId) => ({ appId, deviceId }))

export const START_DEVICE_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_DEVICE_EVENT_STREAM_SUCCESS =
  createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_DEVICE_EVENT_STREAM_FAILURE =
  createStartEventsStreamFailureActionType(SHARED_NAME)
export const STOP_DEVICE_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)

export const PAUSE_DEVICE_EVENT_STREAM = createPauseEventsStreamActionType(SHARED_NAME)

export const RESUME_DEVICE_EVENT_STREAM = createResumeEventsStreamActionType(SHARED_NAME)

export const CLEAR_DEVICE_EVENTS = createClearEventsActionType(SHARED_NAME)

export const SET_DEVICE_EVENTS_FILTER = createSetEventsFilterActionType(SHARED_NAME)

export const GET_DEVICE_EVENT_MESSAGE_SUCCESS = createGetEventMessageSuccessActionType(SHARED_NAME)

export const startDeviceEventsStream = startEventsStream(SHARED_NAME)

export const startDeviceEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)

export const startDeviceEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)

export const pauseDeviceEventsStream = pauseEventsStream(SHARED_NAME)

export const resumeDeviceEventsStream = resumeEventsStream(SHARED_NAME)

export const stopDeviceEventsStream = stopEventsStream(SHARED_NAME)

export const clearDeviceEventsStream = clearEvents(SHARED_NAME)

export const setDeviceEventsFilter = setEventsFilter(SHARED_NAME)
