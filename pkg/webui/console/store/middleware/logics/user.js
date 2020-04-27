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

import api from '@console/api'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import { isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'
import { selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'

import * as accessToken from '@console/lib/access-token'

import * as user from '@console/store/actions/user'

const logoutSequence = async () => {
  const response = await api.console.logout()
  accessToken.clear()
  window.location = response.data.op_logout_uri
}

export default [
  createRequestLogic({
    type: user.LOGOUT,
    async process() {
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
  }),
]
