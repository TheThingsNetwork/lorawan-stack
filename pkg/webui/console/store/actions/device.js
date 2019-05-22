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

export const SHARED_NAME = 'DEVICE'

export const GET_DEV = 'GET_DEVICE'
export const UPDATE_DEV = 'UPDATE_DEVICE'
export const GET_DEV_SUCCESS = 'GET_DEVICE_SUCCESS'
export const GET_DEV_FAILURE = 'GET_DEVICE_FAILURE'
export const START_DEVICE_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_DEVICE_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_DEVICE_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const STOP_DEVICE_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_DEVICE_EVENTS = createClearEventsActionType(SHARED_NAME)

export const getDevice = (appId, deviceId, selector, options) => (
  { type: GET_DEV, appId, deviceId, selector, options }
)

export const updateDevice = (appId, deviceId, patch) => (
  { type: UPDATE_DEV, appId, deviceId, patch }
)

export const getDeviceSuccess = device => (
  { type: GET_DEV_SUCCESS, device }
)

export const getDeviceFailure = error => (
  { type: GET_DEV_FAILURE, error }
)

export const startDeviceEventsStream = startEventsStream(SHARED_NAME)

export const startDeviceEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)

export const startDeviceEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)

export const stopDeviceEventsStream = stopEventsStream(SHARED_NAME)

export const clearDeviceEventsStream = clearEvents(SHARED_NAME)
