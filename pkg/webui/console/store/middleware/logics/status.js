// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import axios from 'axios'

import ONLINE_STATUS from '@ttn-lw/constants/online-status'

import { selectIsConfig } from '@ttn-lw/lib/selectors/env'
import { isNetworkError, isTimeoutError } from '@ttn-lw/lib/errors/utils'

import * as status from '@console/store/actions/status'

import { selectIsOnlineStatus, selectIsOfflineStatus } from '@console/store/selectors/status'

const isRoot = selectIsConfig().base_url

let interval = 5000
const connectionCheck = dispatch => async () => {
  dispatch(status.attemptReconnect())
}

let periodicCheck

const connectionManagementLogic = createLogic({
  type: status.SET_CONNECTION_STATUS,
  process: async ({ action, getState }, dispatch, done) => {
    if (action.payload.onlineStatus === ONLINE_STATUS.CHECKING) {
      try {
        await axios.get(`${isRoot}/auth_info`, { timeout: 5000 })
        dispatch(status.setOnlineStatus(ONLINE_STATUS.ONLINE))
      } catch (error) {
        // If also a simple GET to the auth_info endpoint fails with a
        // network error, we can be sufficiently sure of having gone offline.
        if (isNetworkError(error) || isTimeoutError(error)) {
          dispatch(status.setOnlineStatus(ONLINE_STATUS.OFFLINE))
        }
      }
    }

    if (action.payload.onlineStatus === ONLINE_STATUS.OFFLINE && navigator.onLine) {
      // If the app went offline, try to reconnect periodically.
      dispatch(status.attemptReconnect())
    }

    done()
  },
})

const connectionCheckLogic = createLogic({
  type: status.ATTEMPT_RECONNECT,
  // Additionally to periodic reconnects, freshly incoming request actions will
  // also trigger reconnection attempts, which is why this acction is throttled
  // to 5 seconds.
  throttle: 5000,
  validate: ({ action, getState }, allow, reject) => {
    if (selectIsOfflineStatus(getState()) && navigator.onLine) {
      return allow(action)
    }
    if (Boolean(periodicCheck)) {
      clearTimeout(periodicCheck)
    }
    reject()
  },
  process: async ({ action, getState }, dispatch, done) => {
    try {
      await axios.get(`${isRoot}/auth_info`, { timeout: 4500 })
      dispatch(status.setOnlineStatus(ONLINE_STATUS.ONLINE))
      dispatch(status.attemptReconnectSuccess())
    } catch (error) {
      dispatch(status.attemptReconnectFailure())
    }

    done()
  },
})

const connectionCheckFailLogic = createLogic({
  type: status.ATTEMPT_RECONNECT_FAILURE,
  warnTimeout: 0,
  process: (_, dispatch) => {
    interval = Math.min(interval * 1.5, 60000)
    console.log(interval)
    periodicCheck = setTimeout(connectionCheck(dispatch), interval)
  },
})

export default [connectionManagementLogic, connectionCheckLogic, connectionCheckFailLogic]
