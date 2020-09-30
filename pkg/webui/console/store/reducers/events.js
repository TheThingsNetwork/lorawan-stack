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

import CONNECTION_STATUS from '@console/constants/connection-status'
import EVENT_STORE_LIMIT from '@console/constants/event-store-limit'

import { getCombinedDeviceId } from '@ttn-lw/lib/selectors/id'

import {
  createStatusReconnectedEvent,
  createStatusClearedEvent,
  createStatusClosedEvent,
} from '@console/lib/events/definitions'
import { createSyntheticEventFromError } from '@console/lib/events/utils'

import {
  createGetEventMessagesSuccessActionType,
  createGetEventMessageFailureActionType,
  createStartEventsStreamActionType,
  createStartEventsStreamSuccessActionType,
  createStartEventsStreamFailureActionType,
  createStopEventsStreamActionType,
  createEventStreamClosedActionType,
  createClearEventsActionType,
} from '@console/store/actions/events'

const addEvent = (events, event) => {
  // See https://github.com/TheThingsNetwork/lorawan-stack/pull/2989
  if (event.name === 'events.stream.start' || event.name === 'events.stream.stop') {
    return events
  }

  const currentEvents = events

  // Keep events sorted in descending order by `time`.
  let insertIndex = 0
  while (insertIndex < currentEvents.length) {
    const currentEventTime = currentEvents[insertIndex].time

    if (event.time < currentEventTime) {
      insertIndex += 1
    } else {
      break
    }
  }

  return [...currentEvents.slice(0, insertIndex), event, ...currentEvents.slice(insertIndex)]
}
const defaultState = {
  events: [],
  error: undefined,
  interrupted: false,
  status: CONNECTION_STATUS.DISCONNECTED,
}

const createNamedEventReducer = function(reducerName = '') {
  const START_EVENTS = createStartEventsStreamActionType(reducerName)
  const START_EVENTS_SUCCESS = createStartEventsStreamSuccessActionType(reducerName)
  const START_EVENTS_FAILURE = createStartEventsStreamFailureActionType(reducerName)
  const STOP_EVENTS = createStopEventsStreamActionType(reducerName)
  const GET_EVENTS_SUCCESS = createGetEventMessagesSuccessActionType(reducerName)
  const GET_EVENT_FAILURE = createGetEventMessageFailureActionType(reducerName)
  const CLEAR_EVENTS = createClearEventsActionType(reducerName)
  const EVENT_STREAM_CLOSED = createEventStreamClosedActionType(reducerName)

  return function(state = defaultState, action) {
    switch (action.type) {
      case START_EVENTS:
        return {
          ...state,
          status: CONNECTION_STATUS.CONNECTING,
        }
      case START_EVENTS_SUCCESS:
        return {
          ...(state.interrupted ? addEvent(state, createStatusReconnectedEvent()) : state),
          status: CONNECTION_STATUS.CONNECTED,
          interrupted: false,
          error: undefined,
        }
      case GET_EVENTS_SUCCESS:
        return addEvents(state, action.events)
      case START_EVENTS_FAILURE:
        return {
          ...(!state.interrupted
            ? addEvent(state, createSyntheticEventFromError(action.error))
            : event.state),
          error: action.error,
          status: CONNECTION_STATUS.DISCONNECTED,
        }
      case GET_EVENT_FAILURE:
        return {
          ...addEvent(state, createSyntheticEventFromError(action.error)),
          status: CONNECTION_STATUS.DISCONNECTED,
          interrupted: true,
        }
      case STOP_EVENTS:
        return {
          ...state,
          status: CONNECTION_STATUS.DISCONNECTED,
          interrupted: false,
        }
      case EVENT_STREAM_CLOSED:
        return {
          ...addEvent(state, createStatusClosedEvent()),
          status: CONNECTION_STATUS.DISCONNECTED,
          interrupted: true,
        }
      case CLEAR_EVENTS:
        return {
          ...state,
          events: [createStatusClearedEvent()],
          truncated: false,
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
  const GET_EVENTS_SUCCESS = createGetEventMessagesSuccessActionType(reducerName)
  const GET_EVENT_FAILURE = createGetEventMessageFailureActionType(reducerName)
  const CLEAR_EVENTS = createClearEventsActionType(reducerName)
  const STOP_EVENTS = createStopEventsStreamActionType(reducerName)
  const EVENT_STREAM_CLOSED = createEventStreamClosedActionType(reducerName)
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
      case EVENT_STREAM_CLOSED:
      case GET_EVENT_FAILURE:
      case GET_EVENTS_SUCCESS:
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
