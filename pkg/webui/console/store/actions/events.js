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

export const createStartEventsStreamActionType = name => `START_${name}_EVENT_STREAM`

export const createStartEventsStreamSuccessActionType = name => `START_${name}_EVENT_STREAM_SUCCESS`

export const createStartEventsStreamFailureActionType = name => `START_${name}_EVENT_STREAM_FAILURE`

export const createStopEventsStreamActionType = name => `STOP_${name}_EVENT_STREAM`

export const createGetEventMessageSuccessActionType = name => `GET_${name}_EVENT_MESSAGE_SUCCESS`

export const createGetEventMessageFailureActionType = name => `GET_${name}_EVENT_MESSAGE_FAILURE`

export const createClearEventsActionType = name => `CLEAR_${name}_EVENTS`

export const startEventsStream = name => id => ({
  type: createStartEventsStreamActionType(name),
  id,
})

export const startEventsStreamSuccess = name => id => ({
  type: createStartEventsStreamSuccessActionType(name),
  id,
})

export const startEventsStreamFailure = name => (id, error) => ({
  type: createStartEventsStreamFailureActionType(name),
  id,
  error,
})

export const stopEventsStream = name => id => ({ type: createStopEventsStreamActionType(name), id })

export const getEventMessageSuccess = name => (id, event) => ({
  type: createGetEventMessageSuccessActionType(name),
  id,
  event,
})

export const getEventMessageFailure = name => (id, error) => ({
  type: createGetEventMessageFailureActionType(name),
  id,
  error,
})

export const clearEvents = name => id => ({ type: createClearEventsActionType(name), id })
