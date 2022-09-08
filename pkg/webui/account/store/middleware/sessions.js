// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import tts from '@account/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as session from '@account/store/actions/sessions'

const getUserSessionsLogic = createRequestLogic({
  type: session.GET_USER_SESSIONS_LIST,
  process: async ({ action }) => {
    const { id, params } = action.payload

    const result = await tts.Sessions.getAllSessions(id, {
      page: params?.page,
      limit: params?.limit,
    })

    return { sessions: result.sessions, sessionsTotalCount: result.totalCount }
  },
})

const deleteUserSessionLogic = createRequestLogic({
  type: session.DELETE_USER_SESSION,
  process: async ({ action }) => {
    const { user, sessionId } = action.payload
    const result = await tts.Sessions.deleteSession(user, sessionId)

    return result
  },
})

export default [getUserSessionsLogic, deleteUserSessionLogic]
