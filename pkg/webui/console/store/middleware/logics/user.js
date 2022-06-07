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

import axios from 'axios'

import tts from '@console/api/tts'

import api from '@console/api'

import * as accessToken from '@ttn-lw/lib/access-token'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import { isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'
import { selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'

import * as user from '@console/store/actions/user'

const logoutSequence = async () => {
  const response = await api.console.logout()
  accessToken.clear()
  window.location = response.data.op_logout_uri
}

const logoutLogic = createRequestLogic({
  type: user.LOGOUT,
  process: async () => {
    try {
      await logoutSequence()
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        // If there was an Unauthenticated Error, it either means that the
        // console client or the OAuth app session is no longer valid.
        // In this situation, it's best to try initializing the OAuth
        // roundtrip again. This might provide a new console session cookie
        // with which the propagated logout can be retried. If not, it can
        // be assumed that both console and OAuth app sessions are already
        // terminated, equalling a logged out state. In that case the request
        // logic will perform a page refresh which will initialize the auth
        // flow again.
        await axios.get(
          `${selectApplicationRootPath()}/login/ttn-stack?next=${window.location.pathname}`,
        )
        await logoutSequence()
      } else {
        throw err
      }
    }
  },
})

const getUserInvitationsLogic = createRequestLogic({
  type: user.GET_USER_INVITATIONS,
  process: async ({ action }) => {
    const {
      params: { page, limit, query, order },
    } = action.payload
    const { selectors } = action.meta

    const data = query
      ? await tts.Users.search(
          {
            page,
            limit,
            query,
            order,
          },
          selectors,
        )
      : await tts.Users.getAllInvitations({ page, limit, order }, selectors)

    return data
  },
})

const sendInviteLogic = createRequestLogic({
  type: user.SEND_INVITE,
  process: async ({ action }) => {
    const { email } = action.payload

    return await tts.Users.sendInvite(email)
  },
})

const deleteInviteLogic = createRequestLogic({
  type: user.DELETE_INVITE,
  process: async ({ action }) => {
    const { email } = action.payload

    return await tts.Users.deleteInvite(email)
  },
})

export default [logoutLogic, getUserInvitationsLogic, sendInviteLogic, deleteInviteLogic]
