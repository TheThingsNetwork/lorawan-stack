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

import { createAction } from 'redux-actions'

import ONLINE_STATUS from '@ttn-lw/constants/online-status'

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const SET_CONNECTION_STATUS = 'SET_CONNECTION_STATUS'

export const setStatusOnline = createAction(SET_CONNECTION_STATUS, (isOnline = true) => ({
  onlineStatus: isOnline ? ONLINE_STATUS.ONLINE : ONLINE_STATUS.OFFLINE,
}))
export const setStatusChecking = createAction(SET_CONNECTION_STATUS, () => ({
  onlineStatus: ONLINE_STATUS.CHECKING,
}))

export const SET_LOGIN_STATUS = 'SET_LOGIN_STATUS'

export const setLoginStatus = createAction(
  SET_LOGIN_STATUS,
  (isLoggedIn, sessionId, sessionExpiresAt) => ({
    isLoggedIn,
    sessionId,
    sessionExpiresAt,
  }),
)

export const ATTEMPT_RECONNECT = 'ATTEMPT_RECONNECT'
export const attemptReconnect = createAction(ATTEMPT_RECONNECT)
export const [
  { success: ATTEMPT_RECONNECT_SUCCESS, failure: ATTEMPT_RECONNECT_FAILURE },
  { success: attemptReconnectSuccess, failure: attemptReconnectFailure },
] = createRequestActions(ATTEMPT_RECONNECT)
