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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationByIdDeleteActions,
  createPaginationByParentRequestActions,
} from '@ttn-lw/lib/store/actions/pagination'

export const GET_MAC_SETTINGS_PROFILES_LIST_BASE = 'GET_MAC_SETTINGS_PROFILES_LIST'
export const [
  {
    request: GET_MAC_SETTINGS_PROFILES_LIST,
    success: GET_MAC_SETTINGS_PROFILES_LIST_SUCCESS,
    failure: GET_MAC_SETTINGS_PROFILES_LIST_FAILURE,
  },
  {
    request: getMacSettingsProfilesList,
    success: getMacSettingsProfilesListSuccess,
    failure: getMacSettingsProfilesListFailure,
  },
] = createPaginationByParentRequestActions('MAC_SETTINGS_PROFILES')

export const CREATE_MAC_SETTINGS_PROFILE_BASE = 'CREATE_MAC_SETTINGS_PROFILE'
export const [
  {
    request: CREATE_MAC_SETTINGS_PROFILE,
    success: CREATE_MAC_SETTINGS_PROFILE_SUCCESS,
    failure: CREATE_MAC_SETTINGS_PROFILE_FAILURE,
  },
  {
    request: createMacSettingsProfile,
    success: createMacSettingsProfileSuccess,
    failure: createMacSettingsProfileFailure,
  },
] = createRequestActions(CREATE_MAC_SETTINGS_PROFILE_BASE, (appId, macSettingsProfile) => ({
  appId,
  macSettingsProfile,
}))

export const UPDATE_MAC_SETTINGS_PROFILE_BASE = 'UPDATE_MAC_SETTINGS_PROFILE'
export const [
  {
    request: UPDATE_MAC_SETTINGS_PROFILE,
    success: UPDATE_MAC_SETTINGS_PROFILE_SUCCESS,
    failure: UPDATE_MAC_SETTINGS_PROFILE_FAILURE,
  },
  {
    request: updateMacSettingsProfile,
    success: updateMacSettingsProfileSuccess,
    failure: updateMacSettingsProfileFailure,
  },
] = createRequestActions(UPDATE_MAC_SETTINGS_PROFILE_BASE, (appId, macSettingsProfile) => ({
  appId,
  macSettingsProfile,
}))

export const DELETE_MAC_SETTINGS_PROFILE_BASE = 'DELETE_MAC_SETTINGS_PROFILE'
export const [
  {
    request: DELETE_MAC_SETTINGS_PROFILE,
    success: DELETE_MAC_SETTINGS_PROFILE_SUCCESS,
    failure: DELETE_MAC_SETTINGS_PROFILE_FAILURE,
  },
  {
    request: deleteMacSettingsProfile,
    success: deleteMacSettingsProfileSuccess,
    failure: deleteMacSettingsProfileFailure,
  },
] = createPaginationByIdDeleteActions('MAC_SETTINGS_PROFILES', (id, targetId) => ({ id, targetId }))
