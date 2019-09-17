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

import { defineMessages } from 'react-intl'

import CreateFetchSelect from '../fetch-select'

import {
  selectGsFrequencyPlans,
  selectNsFrequencyPlans,
  selectFrequencyPlansError,
  selectFrequencyPlansFetching,
} from '../../store/selectors/configuration'

import { getGsFrequencyPlans, getNsFrequencyPlans } from '../../store/actions/configuration'

const m = defineMessages({
  title: 'Frequency Plan',
  warning: 'Could not retrieve the list of available frequency plans',
})

const formatOptions = plans => plans.map(plan => ({ value: plan.id, label: plan.name }))

const CreateFrequencyPlansSelector = source =>
  CreateFetchSelect({
    fetchOptions: source === 'ns' ? getNsFrequencyPlans : getGsFrequencyPlans,
    optionsSelector: source === 'ns' ? selectNsFrequencyPlans : selectGsFrequencyPlans,
    fetchingSelector: selectFrequencyPlansFetching,
    errorSelector: selectFrequencyPlansError,
    defaultWarning: m.warning,
    defaultTitle: m.title,
    optionsFormatter: formatOptions,
  })

export const GsFrequencyPlansSelect = CreateFrequencyPlansSelector('gs')
export const NsFrequencyPlansSelect = CreateFrequencyPlansSelector('ns')
