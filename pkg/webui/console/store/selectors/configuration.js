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

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import {
  GET_NS_FREQUENCY_PLANS_BASE,
  GET_GS_FREQUENCY_PLANS_BASE,
} from '@console/store/actions/configuration'

const EMPTY_ARRAY = []

const selectConfigurationStore = state => state.configuration

export const selectGsFrequencyPlans = state => {
  const store = selectConfigurationStore(state)

  return store.gsFrequencyPlans
}

export const selectNsFrequencyPlans = state => {
  const store = selectConfigurationStore(state)

  return store.nsFrequencyPlans
}

export const selectFrequencyPlansError = createErrorSelector([
  GET_NS_FREQUENCY_PLANS_BASE,
  GET_GS_FREQUENCY_PLANS_BASE,
])

export const selectFrequencyPlansFetching = createFetchingSelector([
  GET_NS_FREQUENCY_PLANS_BASE,
  GET_GS_FREQUENCY_PLANS_BASE,
])

export const selectBandDefinitions = state => {
  const store = selectConfigurationStore(state)

  return store.bandDefinitions
}

export const selectDataRates = (state, bandId, phyVersion) => {
  const bandDefinitions = selectBandDefinitions(state)

  return bandDefinitions[bandId]?.band[phyVersion]?.data_rates || EMPTY_ARRAY
}
