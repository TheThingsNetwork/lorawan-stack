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

import api from '@account/api'

import * as init from '@ttn-lw/lib/store/actions/init'
import { isPermissionDeniedError, isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'
import { promisifyDispatch } from '@ttn-lw/lib/store/middleware/request-promise-middleware'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { setLoginStatus } from '@ttn-lw/lib/store/actions/status'

import * as user from '@account/store/actions/user'

const accountAppInitLogic = createLogic({
  type: init.INITIALIZE,
  process: async (_, dispatch, done) => {
    let sessionId
    try {
      const meResult = await api.account.me()
      // Using `store.dispatch` since redux logic's dispatch won't return
      // the (promisified) action result like regular dispatch does.
      await promisifyDispatch(dispatch)(
        attachPromise(
          user.getUser(meResult.data.user.ids.user_id, [
            'profile_picture',
            'name',
            'description',
            'primary_email_address',
            'admin',
          ]),
        ),
      )
      sessionId = meResult.data.session_id
      dispatch(setLoginStatus(true, sessionId, meResult.data.expires_at))
    } catch (error) {
      if (!isUnauthenticatedError(error) && !isPermissionDeniedError(error)) {
        const initError = error?.data || error
        dispatch(init.initializeFailure(initError))
        return done()
      }
      // Unauthenticated or forbidden errors mean that the user is logged out.
      // This is expected and should not make the initialization fail.
    }

    // eslint-disable-next-line no-console
    console.log('Account app initialization successful!')
    dispatch(init.initializeSuccess(sessionId))
    done()
  },
})

export default accountAppInitLogic
