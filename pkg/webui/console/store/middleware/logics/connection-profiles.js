// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as connectionProfiles from '@console/store/actions/connection-profiles'

const getConnectionProfilesLogic = createRequestLogic({
  type: connectionProfiles.GET_CONNECTION_PROFILES_LIST,
  process: async ({ action }) => {
    const { type } = action.payload

    // TODO: Change call to fetch connection profiles
    const res = {
      profiles: [],
      totalCount: 0,
    }

    return { entities: res.profiles, profilesTotalCount: res.totalCount }
  },
})

const deleteConnectionProfileLogic = createRequestLogic({
  type: connectionProfiles.DELETE_CONNECTION_PROFILE,
  process: async ({ action }) => {
    const { id } = action.payload

    // TODO: Change call to delete connection profiles
    // await tts.Authorizations.deleteToken(userId, clientId, id)

    return { id }
  },
})

export default [getConnectionProfilesLogic, deleteConnectionProfileLogic]
