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

import { handleActions } from 'redux-actions'

import { GET_DEFAULT_MAC_SETTINGS_SUCCESS } from '@console/store/actions/network-server'

import { generateDefaultMacSettingsKey } from '@console/store/selectors/network-server'

const defaultState = {
  defaultMacSettings: {},
}

export default handleActions(
  {
    [GET_DEFAULT_MAC_SETTINGS_SUCCESS]: (
      state,
      { payload: { frequencyPlanId, lorawanPhyVersion, defaultMacSettings } },
    ) => ({
      ...state,
      defaultMacSettings: {
        ...state.defaultMacSettings,
        [generateDefaultMacSettingsKey(frequencyPlanId, lorawanPhyVersion)]: defaultMacSettings,
      },
    }),
  },
  defaultState,
)
