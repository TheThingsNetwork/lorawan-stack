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

import api from '@console/api'

import { clear as clearAccessToken } from '@console/lib/access-token'

import * as user from '@console/store/actions/user'
import * as init from '@console/store/actions/init'

import createRequestLogic from './lib'

const consoleAppLogic = createRequestLogic({
  type: init.INITIALIZE,
  async process(_, dispatch) {
    dispatch(user.getUserRights())

    let info, rights

    try {
      // There is no way to retrieve the current user directly within the
      // console app, so first get the authentication info and only afterwards
      // fetch the user.
      info = await api.users.authInfo()
      rights = info.oauth_access_token.rights
      dispatch(user.getUserRightsSuccess(rights))
    } catch (error) {
      if (error.code === 16) {
        // The access token was not found, so we can delete it from local
        // storage to obtain a new one.
        clearAccessToken()
      }
      dispatch(user.getUserRightsFailure())
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
      } catch (error) {
        dispatch(user.getUserMeFailure(error))
      }
    }

    // eslint-disable-next-line no-console
    console.log('Console initialization successful!')

    return
  },
})

export default [consoleAppLogic]
