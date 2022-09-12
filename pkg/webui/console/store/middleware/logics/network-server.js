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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as ns from '@console/store/actions/network-server'

import { selectDefaultMacSettings } from '@console/store/selectors/network-server'

const getDefaultMacSettings = createRequestLogic({
  type: ns.GET_DEFAULT_MAC_SETTINGS,
  process: async ({ getState, action }) => {
    const { frequencyPlanId, lorawanPhyVersion } = action.payload
    const state = getState()
    const cachedResult = selectDefaultMacSettings(state, frequencyPlanId, lorawanPhyVersion)

    // Default MAC settings typically don't change so we can
    // cache the result until the next refresh.
    const defaultMacSettings =
      cachedResult || (await tts.Ns.getDefaultMacSettings(frequencyPlanId, lorawanPhyVersion))

    return {
      defaultMacSettings,
      frequencyPlanId,
      lorawanPhyVersion,
    }
  },
})

export default [getDefaultMacSettings]