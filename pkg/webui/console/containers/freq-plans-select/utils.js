// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

export const formatOptions = plans =>
  plans.map(plan => ({ value: plan.id, label: plan.name, tag: plan.band_id }))
export const m = defineMessages({
  warning: 'Frequency plans unavailable',
  none: 'Do not set a frequency plan',
  selectFrequencyPlan: 'Select a frequency plan...',
  addFrequencyPlan: 'Add frequency plan',
  frequencyPlanDescription:
    'Note: most gateways use a single frequency plan. Some 16 and 64 channel gateways however allow setting multiple within the same band.',
})
