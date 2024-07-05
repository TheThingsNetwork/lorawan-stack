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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const GET_CONNECTION_PROFILES_LIST_BASE = 'GET_CONNECTION_PROFILES_LIST'
export const [
  {
    request: GET_CONNECTION_PROFILES_LIST,
    success: GET_CONNECTION_PROFILES_LIST_SUCCESS,
    failure: GET_CONNECTION_PROFILES_LIST_FAILURE,
  },
  {
    request: getConnectionProfilesList,
    success: getConnectionProfilesListSuccess,
    failure: getConnectionProfilesListFailure,
  },
] = createRequestActions(
  GET_CONNECTION_PROFILES_LIST_BASE,
  type => ({ type }),
  (type, selector) => ({ selector }),
)

export const DELETE_CONNECTION_PROFILE_BASE = 'DELETE_CONNECTION_PROFILE'
export const [
  {
    request: DELETE_CONNECTION_PROFILE,
    success: DELETE_CONNECTION_PROFILE_SUCCESS,
    failure: DELETE_CONNECTION_PROFILE_FAILURE,
  },
  {
    request: deleteConnectionProfile,
    success: deleteConnectionProfileSuccess,
    failure: deleteConnectionProfileFailure,
  },
] = createRequestActions(DELETE_CONNECTION_PROFILE_BASE, (type, id) => ({
  type,
  id,
}))
