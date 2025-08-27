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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import {
  CREATE_MAC_SETTINGS_PROFILE,
  GET_MAC_SETTINGS_PROFILES_LIST,
} from '@console/store/actions/mac-settings-profiles'

const getMacSettingsProfiles = createRequestLogic({
  type: GET_MAC_SETTINGS_PROFILES_LIST,
  process: async ({ action }) => {
    const {
      parentId,
      params: { page, limit, order },
    } = action.payload

    const data = await tts.NsMACSettingsProfiles.getAll(parentId, { page, limit, order })

    return { entities: data.mac_settings_profiles, totalCount: data.totalCount }
  },
})

const createMacSettingsProfile = createRequestLogic({
  type: CREATE_MAC_SETTINGS_PROFILE,
  process: async ({ action }) => {
    const { appId, macSettingsProfile } = action.payload
    const profile = await tts.NsMACSettingsProfiles.create(appId, macSettingsProfile)

    return profile
  },
})

export default [getMacSettingsProfiles, createMacSettingsProfile]
