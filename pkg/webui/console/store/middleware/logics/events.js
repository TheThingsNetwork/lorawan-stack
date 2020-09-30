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

import { createLogic } from 'redux-logic'

import CONNECTION_STATUS from '@console/constants/connection-status'

import createBufferedProcess from '@ttn-lw/lib/create-buffered-process'
import { getCombinedDeviceId } from '@ttn-lw/lib/selectors/id'
import { isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'

import {
  createEventsStatusSelector,
  createEventsInterruptedSelector,
  createInterruptedStreamsSelector,
} from '@console/store/selectors/events'
import { selectConnectionStatus } from '@console/store/selectors/status'

import {
  createStartEventsStreamActionType,
  createStopEventsStreamActionType,
  createStartEventsStreamFailureActionType,
  createStartEventsStreamSuccessActionType,
  createEventStreamClosedActionType,
  createGetEventMessageFailureActionType,
  createGetEventMessagesSuccessActionType,
  getEventMessagesSuccess,
  getEventMessageFailure,
  startEventsStreamFailure,
  startEventsStreamSuccess,
  eventStreamClosed,
  startEventsStream,
} from '../../actions/events'
import { SET_CONNECTION_STATUS } from '../../actions/status'

/**
 * Creates `redux-logic` logic from processing entity events.
 *
 * @param {string} reducerName - The name of an entity used to create the events reducer.
 * @param {string} entityName - The name of an entity.
 * @param {Function} onEventsStart - A function to be called to start the events stream.
 * Should accept a list of entity ids.
 * @returns {object} - The `redux-logic` (decorated) logic.
 */
const createEventsConnectLogics = function(reducerName, entityName, onEventsStart) {
  const START_EVENTS = createStartEventsStreamActionType(reducerName)
  const START_EVENTS_SUCCESS = createStartEventsStreamSuccessActionType(reducerName)
  const START_EVENTS_FAILURE = createStartEventsStreamFailureActionType(reducerName)
  const STOP_EVENTS = createStopEventsStreamActionType(reducerName)
  const EVENT_STREAM_CLOSED = createEventStreamClosedActionType(reducerName)
  const GET_EVENT_MESSAGE_FAILURE = createGetEventMessageFailureActionType(reducerName)
  const GET_EVENT_MESSAGES_SUCCESS = createGetEventMessagesSuccessActionType(reducerName)
  const startEventsSuccess = startEventsStreamSuccess(reducerName)
  const startEventsFailure = startEventsStreamFailure(reducerName)
  const closeEvents = eventStreamClosed(reducerName)
  const startEvents = startEventsStream(reducerName)
  const getEventsSuccess = getEventMessagesSuccess(reducerName)
  const getEventFailure = getEventMessageFailure(reducerName)
  const selectEntityEventsStatus = createEventsStatusSelector(entityName)
  const selectEntityEventsInterrupted = createEventsInterruptedSelector(entityName)
  const selectInterruptedStreams = createInterruptedStreamsSelector(entityName)

  let channel = null

  return [
    createLogic({
      type: START_EVENTS,
      cancelType: [STOP_EVENTS, START_EVENTS_FAILURE],
      warnTimeout: 0,
      processOptions: {
        dispatchMultiple: true,
      },
      validate({ getState, action = {} }, allow, reject) {
        if (!action.id) {
          reject()
          return
        }

        const id = typeof action.id === 'object' ? getCombinedDeviceId(action.id) : action.id

        // Only proceed if not already connected and online.
        const state = getState()
        const isOnline = selectConnectionStatus(state)
        const status = selectEntityEventsStatus(state, id)
        const connected = status === CONNECTION_STATUS.CONNECTED
        const connecting = status === CONNECTION_STATUS.CONNECTING
        if (connected || connecting || !isOnline) {
          reject()
          return
        }

        allow(action)
      },
      async process({ getState, action }, dispatch) {
        const { id } = action

        const {
          addToBuffer: addToEventBuffer,
          clearBuffer: clearEventBuffer,
        } = createBufferedProcess(events => {
          const processedEvents = events
            .filter(
              // See https://github.com/TheThingsNetwork/lorawan-stack/pull/2989
              event => event.name !== 'events.stream.start' && event.name !== 'events.stream.stop',
            )
            .sort((a, b) => (a.time > b.time ? -1 : 1))

          if (processedEvents.length > 0) {
            dispatch(getEventsSuccess(id, processedEvents))
          }
        })

        try {
          channel = await onEventsStart([id])

          channel.on('start', () => dispatch(startEventsSuccess(id)))
          channel.on('chunk', addToEventBuffer)
          channel.on('error', error => {
            // Clear event buffer before committing the failure event
            // to avoid race conditions.
            clearEventBuffer()
            dispatch(getEventFailure(id, error))
          })
          channel.on('close', () => dispatch(closeEvents(id)))
        } catch (error) {
          if (isUnauthenticatedError(error)) {
            // The user is no longer authenticated; reinitiate the auth flow
            // by refreshing the page.
            window.location.reload()
          } else {
            dispatch(startEventsFailure(id, error))
          }
        }
      },
    }),
    createLogic({
      type: [STOP_EVENTS, START_EVENTS_FAILURE],
      validate({ getState, action = {} }, allow, reject) {
        if (!action.id) {
          reject()
          return
        }

        const id = typeof action.id === 'object' ? getCombinedDeviceId(action.id) : action.id

        // Only proceed if connected.
        const status = selectEntityEventsStatus(getState(), id)
        const connected = status === CONNECTION_STATUS.CONNECTED
        const connecting = status === CONNECTION_STATUS.CONNECTING
        if (!connected && !connecting) {
          reject()
          return
        }

        allow(action)
      },
      process(_, __, done) {
        if (channel) {
          channel.close()
        }
        done()
      },
    }),
    createLogic({
      type: [GET_EVENT_MESSAGE_FAILURE, EVENT_STREAM_CLOSED],
      cancelType: [START_EVENTS_SUCCESS, GET_EVENT_MESSAGES_SUCCESS, STOP_EVENTS],
      warnTimeout: 0,
      validate({ getState, action = {} }, allow, reject) {
        if (!action.id) {
          reject()
          return
        }

        const id = typeof action.id === 'object' ? getCombinedDeviceId(action.id) : action.id

        // Only proceed if connected and not interrupted.
        const status = selectEntityEventsStatus(getState(), id)
        const connected = status === CONNECTION_STATUS.CONNECTED
        const interrupted = selectEntityEventsInterrupted(getState(), id)
        if (!connected && interrupted) {
          reject()
        }

        allow(action)
      },
      process({ getState, action }, dispatch, done) {
        const isOnline = selectConnectionStatus(getState())

        // If the app is not offline, try to reconnect periodically.
        if (isOnline) {
          const reconnector = setInterval(() => {
            // Only proceed if still disconnected, interrupted and online.
            const state = getState()
            const id = typeof action.id === 'object' ? getCombinedDeviceId(action.id) : action.id
            const status = selectEntityEventsStatus(state, id)
            const disconnected = status === CONNECTION_STATUS.DISCONNECTED
            const interrupted = selectEntityEventsInterrupted(state, id)
            const isOnline = selectConnectionStatus(state)
            if (disconnected && interrupted && isOnline) {
              dispatch(startEvents(id))
            } else {
              clearInterval(reconnector)
              done()
            }
          }, 5000)
        } else {
          done()
        }
      },
    }),
    createLogic({
      type: SET_CONNECTION_STATUS,
      process({ getState, action }, dispatch, done) {
        const isOnline = action.payload.isOnline

        if (isOnline) {
          const state = getState()
          for (const id in selectInterruptedStreams(state)) {
            const status = selectEntityEventsStatus(state, id)
            const disconnected = status === CONNECTION_STATUS.DISCONNECTED

            // If the app reconnected to the internet and there is a pending
            // interrupted stream connection, try to reconnect.
            if (disconnected) {
              dispatch(dispatch(startEvents(id)))
            }
          }
        }

        done()
      },
    }),
  ]
}

export default createEventsConnectLogics
