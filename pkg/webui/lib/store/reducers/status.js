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

import ONLINE_STATUS from '@ttn-lw/constants/online-status'

import { SET_CONNECTION_STATUS, SET_LOGIN_STATUS } from '@ttn-lw/lib/store/actions/status'

const defaultState = {
  onlineStatus: ONLINE_STATUS.ONLINE,
  isLoggedIn: false,
  sessionId: undefined,
  sessionExpiresAt: undefined,
}

const status = (state = defaultState, { type, payload }) => {
  switch (type) {
    case SET_CONNECTION_STATUS:
      return {
        ...state,
        onlineStatus: payload.onlineStatus,
      }
    case SET_LOGIN_STATUS:
      return {
        ...state,
        isLoggedIn: payload.isLoggedIn,
        sessionId: payload.sessionId,
        sessionExpiresAt: payload.sessionExpiresAt,
      }
    default:
      return state
  }
}

export default status
