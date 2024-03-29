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

import CreateFetchSelect from '@console/containers/fetch-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getGsFrequencyPlans, getNsFrequencyPlans } from '@console/store/actions/configuration'

import {
  selectGsFrequencyPlans,
  selectNsFrequencyPlans,
  selectFrequencyPlansError,
  selectFrequencyPlansFetching,
} from '@console/store/selectors/configuration'

import { formatOptions, m } from './utils'

export const CreateFrequencyPlansSelect = (source, options = {}) =>
  CreateFetchSelect({
    fetchOptions: source === 'ns' ? getNsFrequencyPlans : getGsFrequencyPlans,
    optionsSelector: source === 'ns' ? selectNsFrequencyPlans : selectGsFrequencyPlans,
    fetchingSelector: selectFrequencyPlansFetching,
    errorSelector: selectFrequencyPlansError,
    defaultWarning: m.warning,
    defaultTitle: sharedMessages.frequencyPlan,
    optionsFormatter: formatOptions,
    additionalOptions: source === 'gs' ? [{ value: 'no-frequency-plan', label: m.none }] : [],
    ...options,
  })

export const GsFrequencyPlansSelect = CreateFrequencyPlansSelect('gs')
export const NsFrequencyPlansSelect = CreateFrequencyPlansSelect('ns')
