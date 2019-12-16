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

import { defineMessages } from 'react-intl'

import * as user from '../../actions/user'
import * as init from '../../actions/init'
import api from '../../../api'
import createRequestLogic from './lib'

const m = defineMessages({
  errTooFewRights: 'Your account does not possess sufficient rights to use the console.',
  errStateRequested:
    'Your account still needs to be approved by an administrator. You will receive a confirmation email once your account is approved.',
  errStateRejected: 'Your account has been rejected by an administrator.',
  errStateSuspended:
    'Your account has been suspended by an administrator. Please contact support for further information about your account status.',
  errEmailValidation: 'Your account is restricted until your email address has been validated.',
})

// Define a minimum set of rights, without which it makes no sense to use the
// console
const minimumRights = ['RIGHT_APPLICATION', 'RIGHT_GATEWAY', 'RIGHT_ORGANIZATION']

const consoleAppLogic = createRequestLogic({
  type: init.INITIALIZE,
  async process(_, dispatch) {
    dispatch(user.getUserRights())

    let info, rights

    try {
      // there is no way to retrieve the current user directly
      // within the console app, so first get the authentication info
      // and only afterwards fetch the user
      info = await api.users.authInfo()
      rights = info.oauth_access_token.rights
      dispatch(user.getUserRightsSuccess(rights))
    } catch (error) {
      dispatch(user.getUserRightsFailure())
      dispatch(user.logout())
      info = undefined
    }

    if (info) {
      try {
        dispatch(user.getUserMe())
        const userId = info.oauth_access_token.user_ids.user_id
        const userResult = await api.users.get(userId, [
          'state',
          'name',
          'primary_email_address_validated_at',
        ])
        userResult.isAdmin = info.is_admin || false
        dispatch(user.getUserMeSuccess(userResult))

        // Check whether the user account has sufficient rights to use the
        // console
        if (!info.is_admin && !rights.some(r => minimumRights.some(mr => r.startsWith(mr)))) {
          // Provide relevant error messages if possible
          if (userResult.state === 'STATE_REQUESTED') {
            throw m.errStateRequested
          } else if (userResult.state === 'STATE_REJECTED') {
            throw m.errStateRejected
          } else if (userResult.state === 'STATE_SUSPENDED') {
            throw m.errStateSuspended
          } else if (!userResult.primary_email_address_validated_at) {
            throw m.errEmailValidation
          }

          throw m.errTooFewRights
        }
      } catch (error) {
        dispatch(user.getUserMeFailure(error))
        dispatch(user.logout())
      }
    }

    // eslint-disable-next-line no-console
    console.log('Console initialization successful!')

    return
  },
})

export default [consoleAppLogic]
