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
import api from '../../../api'
import {
  createStartEventsStreamActionType,
  createStopEventsStreamActionType,
  createStartEventsStreamFailureActionType,
  getEventMessageSuccess,
  getEventMessageFailure,
  startEventsStreamFailure,
  startEventsStreamSuccess,
  stopEventsStream,
} from '../../actions/events'
import { selectApplicationEventsStatus } from '../../selectors/application'

const createEventsConnectLogics = function (name, entity) {
  const START_EVENTS = createStartEventsStreamActionType(name)
  const START_EVENTS_FAILURE = createStartEventsStreamFailureActionType(name)
  const STOP_EVENTS = createStopEventsStreamActionType(name)
  const startEventsSuccess = startEventsStreamSuccess(name)
  const startEventsFailure = startEventsStreamFailure(name)
  const stopEvents = stopEventsStream(name)
  const getEventSuccess = getEventMessageSuccess(name)
  const getEventFailure = getEventMessageFailure(name)

  let channel = null

  return [
    createLogic({
      type: START_EVENTS,
      warnTimeout: 0,
      validate ({ getState, action }, allow, reject) {
        const { id } = action
        if (!id) {
          reject()
        }

        // only proceed if not already connected
        const status = selectApplicationEventsStatus(getState(), { id })
        const connected = status === CONNECTION_STATUS.CONNECTED
        const connecting = status === CONNECTION_STATUS.CONNECTING
        if (connected || connecting) {
          reject()
        }

        allow(action)
      },
      async process ({ action }, dispatch, done) {
        const { eventsSubscribe } = api[entity]
        const { id } = action

        try {
          channel = await eventsSubscribe([ id ])
          channel.on('start', () => dispatch(startEventsSuccess(id)))
          channel.on('event', message => dispatch(getEventSuccess(id, message)))
          channel.on('error', error => dispatch(getEventFailure(id, error)))
          channel.on('close', () => dispatch(stopEvents(id)))
        } catch (error) {
          dispatch(startEventsFailure(error))
          done()
        }
      },
    }),
    createLogic({
      type: [ STOP_EVENTS, START_EVENTS_FAILURE ],
      validate ({ getState, action }, allow, reject) {
        const { id } = action
        if (!id || !channel) {
          reject()
        }

        // only proceed if connected
        const status = selectApplicationEventsStatus(getState(), { id })
        const disconnected = status === CONNECTION_STATUS.DISCONNECTED
        const unknown = status === CONNECTION_STATUS.UNKNOWN
        if (disconnected || unknown) {
          reject()
        }

        allow(action)
      },
      process (helpers, dispatch, done) {
        channel.close()
        done()
      },
    }),
  ]
}

export default createEventsConnectLogics
