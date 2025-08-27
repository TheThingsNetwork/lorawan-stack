// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

import { GET_MAC_SETTINGS_PROFILES_LIST_SUCCESS } from '@console/store/actions/mac-settings-profiles'

const defaultState = {
  entities: {},
  totalCount: null,
  selectedMacProfile: null,
}

const macSettingsProfile = (state = {}, macSettingsProfile) => ({
  ...state,
  ...macSettingsProfile,
})

const macSettingsProfiles = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_MAC_SETTINGS_PROFILES_LIST_SUCCESS:
      const profiles = payload.entities.reduce(
        (acc, c) => {
          const id = c.ids.profile_id

          acc[id] = macSettingsProfile(acc[id], c)
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities: profiles,
        totalCount: payload.totalCount,
      }
    default:
      return state
  }
}

export default macSettingsProfiles
