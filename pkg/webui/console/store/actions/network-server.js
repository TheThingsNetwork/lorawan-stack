// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

export const GET_DEFAULT_MAC_SETTINGS_BASE = 'GET_DEFAULT_MAC_SETTINGS'
export const [
  {
    request: GET_DEFAULT_MAC_SETTINGS,
    success: GET_DEFAULT_MAC_SETTINGS_SUCCESS,
    failure: GET_DEFAULT_MAC_SETTINGS_FAILURE,
  },
  {
    request: getDefaultMacSettings,
    success: getDefaultMacSettingsSuccess,
    failure: getDefaultMacSettingsFailure,
  },
] = createRequestActions(GET_DEFAULT_MAC_SETTINGS_BASE, (frequencyPlanId, lorawanPhyVersion) => ({
  frequencyPlanId,
  lorawanPhyVersion,
}))
