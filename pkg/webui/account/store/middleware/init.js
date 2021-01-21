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

import store from '@account/store'

import api from '@account/api'

import { isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import * as init from '@account/store/actions/init'
import * as user from '@account/store/actions/user'

const accountAppInitLogic = createLogic({
  type: init.INITIALIZE,
  process: async ({ getState, action }, dispatch, done) => {
    try {
      const meResult = await api.account.me()
      try {
        // Using `store.dispatch` since redux logic's dispatch won't return
        // the (promisified) action result like regular dispatch does.
        await store.dispatch(
          attachPromise(
            user.getUser(meResult.data.user.ids.user_id, [
              'profile_picture',
              'name',
              'description',
              'primary_email_address',
            ]),
          ),
        )
      } catch (error) {
        // An error here means that the user is logged out. This does not
        // need to be handled or result in the initialization to fail.
      }
    } catch (error) {
      if (!isUnauthenticatedError(error)) {
        const initError = error.data ? error.data : error
        dispatch(init.initializeFailure(initError))
        return done()
      }
    }

    // eslint-disable-next-line no-console
    console.log('Account app initialization successful!')
    dispatch(init.initializeSuccess())
    done()
  },
})

export default accountAppInitLogic
