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

import tts from '@console/api/tts'

import { clear as clearAccessToken } from '@ttn-lw/lib/access-token'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import * as init from '@ttn-lw/lib/store/actions/init'
import { TokenError } from '@ttn-lw/lib/errors/custom-errors'
import { isPermissionDeniedError, isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'

import * as user from '@console/store/actions/logout'
import { getInboxNotifications, getUnseenNotifications } from '@console/store/actions/notifications'

const consoleAppLogic = createRequestLogic({
  type: init.INITIALIZE,
  noCancelOnRouteChange: true,
  process: async (_, dispatch) => {
    dispatch(user.getUserRights())

    let info, rights

    try {
      // There is no way to retrieve the current user directly within the
      // console app, so first get the authentication info and only afterwards
      // fetch the user.
      info = await tts.Auth.getAuthInfo()
      rights = info.oauth_access_token.rights
      dispatch(user.getUserRightsSuccess(rights))
    } catch (error) {
      if (
        error instanceof TokenError
          ? !isUnauthenticatedError(error?.cause) && !isPermissionDeniedError(error?.cause)
          : !isUnauthenticatedError(error)
      ) {
        throw error
      }

      // Clear existing access token since it does
      // not appear to be valid anymore.
      clearAccessToken()
      dispatch(user.getUserRightsFailure())
      info = undefined
    }

    if (info) {
      try {
        dispatch(user.getUserMe())
        const userId = info.oauth_access_token.user_ids.user_id
        const userResult = await tts.Users.getById(userId, [
          'state',
          'name',
          'primary_email_address_validated_at',
          'profile_picture',
        ])
        userResult.isAdmin = info.is_admin || false
        dispatch(user.getUserMeSuccess(userResult))
        dispatch(getInboxNotifications({ page: 1, limit: 3 }))
        dispatch(getUnseenNotifications(userId))
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
