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

import CONNECTION_STATUS from '../../../constants/connection-status'
import {
  createStartEventsStreamActionType,
  createStopEventsStreamActionType,
  createStartEventsStreamFailureActionType,
  createGetEventMessageFailureActionType,
  getEventMessageSuccess,
  getEventMessageFailure,
  startEventsStreamFailure,
  startEventsStreamSuccess,
  stopEventsStream,
} from '../../actions/events'
import { createEventsStatusSelector } from '../../selectors/events'

/**
 * Creates `redux-logic` logic from processing entity events.
 * @param {string} reducerName - The name of an entity used to create the events reducer.
 * @param {string} entityName - The name of an entity.
 * @param {Function} onEventsStart - A function to be called to start the events stream.
 * Should accept a list of entity ids.
 * @returns {Object} - The `redux-logic` (decorated) logic.
 */
const createEventsConnectLogics = function (
  reducerName,
  entityName,
  onEventsStart,
) {
  const START_EVENTS = createStartEventsStreamActionType(reducerName)
  const START_EVENTS_FAILURE = createStartEventsStreamFailureActionType(reducerName)
  const STOP_EVENTS = createStopEventsStreamActionType(reducerName)
  const GET_EVENT_MESSAGE_FAILURE = createGetEventMessageFailureActionType(reducerName)
  const startEventsSuccess = startEventsStreamSuccess(reducerName)
  const startEventsFailure = startEventsStreamFailure(reducerName)
  const stopEvents = stopEventsStream(reducerName)
  const getEventSuccess = getEventMessageSuccess(reducerName)
  const getEventFailure = getEventMessageFailure(reducerName)
  const selectEntityEventsStatus = createEventsStatusSelector(entityName)

  let channel = null

  return [
    createLogic({
      type: START_EVENTS,
      cancelType: [ STOP_EVENTS, START_EVENTS_FAILURE, GET_EVENT_MESSAGE_FAILURE ],
      warnTimeout: 0,
      processOptions: {
        dispatchMultiple: true,
      },
      validate ({ getState, action }, allow, reject) {
        const { id } = action
        if (!id) {
          reject()
          return
        }

        // only proceed if not already connected
        const status = selectEntityEventsStatus(getState(), id)
        const connected = status === CONNECTION_STATUS.CONNECTED
        const connecting = status === CONNECTION_STATUS.CONNECTING
        if (connected || connecting) {
          reject()
          return
        }

        allow(action)
      },
      async process ({ getState, action }, dispatch) {
        const { id } = action

        try {
          channel = await onEventsStart([ id ])

          channel.on('start', () => dispatch(startEventsSuccess(id)))
          channel.on('event', message => dispatch(getEventSuccess(id, message)))
          channel.on('error', error => dispatch(getEventFailure(id, error)))
          channel.on('close', () => dispatch(stopEvents(id)))
        } catch (error) {
          dispatch(startEventsFailure(id, error))
        }
      },
    }),
    createLogic({
      type: [ STOP_EVENTS, START_EVENTS_FAILURE, GET_EVENT_MESSAGE_FAILURE ],
      validate ({ getState, action }, allow, reject) {
        const { id } = action
        if (!id) {
          reject()
          return
        }

        // only proceed if connected
        const status = selectEntityEventsStatus(getState(), id)
        const disconnected = status === CONNECTION_STATUS.DISCONNECTED
        const unknown = status === CONNECTION_STATUS.UNKNOWN
        if (disconnected || unknown) {
          reject()
          return
        }

        allow(action)
      },
      process (_, __, done) {
        if (channel) {
          channel.close()
        }
        done()
      },
    }),
  ]
}

export default createEventsConnectLogics
