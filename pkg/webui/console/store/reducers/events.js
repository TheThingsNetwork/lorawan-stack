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

import { getCombinedDeviceId } from '@ttn-lw/lib/selectors/id'

import {
  createGetEventMessageSuccessActionType,
  createGetEventMessageFailureActionType,
  createStartEventsStreamActionType,
  createStartEventsStreamSuccessActionType,
  createStartEventsStreamFailureActionType,
  createStopEventsStreamActionType,
  createClearEventsActionType,
} from '@console/store/actions/events'

import CONNECTION_STATUS from '../../constants/connection-status'

const addEvent = (state, event) => {
  const currentEvents = state.events
  const eventTime = new Date(event.time).getTime()

  // Keep events sorted in descending order by `time`.
  let insertIndex = 0
  while (insertIndex < currentEvents.length) {
    const currentEventTime = new Date(currentEvents[insertIndex].time).getTime()

    if (eventTime < currentEventTime) {
      insertIndex += 1
    } else {
      break
    }
  }

  const events = [
    ...currentEvents.slice(0, insertIndex),
    event,
    ...currentEvents.slice(insertIndex),
  ]

  return {
    ...state,
    events,
  }
}
const defaultState = {
  events: [],
  error: undefined,
  status: CONNECTION_STATUS.DISCONNECTED,
}

const createNamedEventReducer = function(reducerName = '') {
  const START_EVENTS = createStartEventsStreamActionType(reducerName)
  const START_EVENTS_SUCCESS = createStartEventsStreamSuccessActionType(reducerName)
  const START_EVENTS_FAILURE = createStartEventsStreamFailureActionType(reducerName)
  const STOP_EVENTS = createStopEventsStreamActionType(reducerName)
  const GET_EVENT_SUCCESS = createGetEventMessageSuccessActionType(reducerName)
  const GET_EVENT_FAILURE = createGetEventMessageFailureActionType(reducerName)
  const CLEAR_EVENTS = createClearEventsActionType(reducerName)

  return function(state = defaultState, action) {
    switch (action.type) {
      case START_EVENTS:
        return {
          ...state,
          status: CONNECTION_STATUS.CONNECTING,
        }
      case START_EVENTS_SUCCESS:
        return {
          ...state,
          error: undefined,
          status: CONNECTION_STATUS.CONNECTED,
        }
      case GET_EVENT_SUCCESS:
        return addEvent(state, action.event)
      case START_EVENTS_FAILURE:
      case GET_EVENT_FAILURE:
        return {
          ...state,
          error: action.error,
          status: CONNECTION_STATUS.ERROR,
        }
      case STOP_EVENTS:
        return {
          ...state,
          status: CONNECTION_STATUS.DISCONNECTED,
        }
      case CLEAR_EVENTS:
        return {
          ...state,
          events: [],
        }
      default:
        return state
    }
  }
}

const createNamedEventsReducer = function(reducerName = '') {
  const START_EVENTS = createStartEventsStreamActionType(reducerName)
  const START_EVENTS_SUCCESS = createStartEventsStreamSuccessActionType(reducerName)
  const START_EVENTS_FAILURE = createStartEventsStreamFailureActionType(reducerName)
  const GET_EVENT_SUCCESS = createGetEventMessageSuccessActionType(reducerName)
  const GET_EVENT_FAILURE = createGetEventMessageFailureActionType(reducerName)
  const CLEAR_EVENTS = createClearEventsActionType(reducerName)
  const STOP_EVENTS = createStopEventsStreamActionType(reducerName)
  const event = createNamedEventReducer(reducerName)

  return function(state = {}, action) {
    if (!action.id) {
      return state
    }

    const id = typeof action.id === 'object' ? getCombinedDeviceId(action.id) : action.id

    switch (action.type) {
      case START_EVENTS:
      case START_EVENTS_FAILURE:
      case START_EVENTS_SUCCESS:
      case STOP_EVENTS:
      case GET_EVENT_FAILURE:
      case GET_EVENT_SUCCESS:
      case CLEAR_EVENTS:
        return {
          ...state,
          [id]: event(state[id], action),
        }
      default:
        return state
    }
  }
}

export default createNamedEventsReducer
