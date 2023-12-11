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

import React, { useCallback, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import frequencyPlans from '@console/constants/frequency-plans'

import KeyValueMap from '@ttn-lw/components/key-value-map'
import Select from '@ttn-lw/components/select'
import Form, { useFormContext } from '@ttn-lw/components/form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { getGsFrequencyPlans } from '@console/store/actions/configuration'

import { selectGsFrequencyPlans } from '@console/store/selectors/configuration'

import { formatOptions, m } from './utils'

const isEmptyFrequencyPlan = value => value?.includes(frequencyPlans.EMPTY_FREQ_PLAN)

const GatewayFrequencyPlansSelect = () => {
  const { values } = useFormContext()
  const { frequency_plan_ids } = values
  const dispatch = useDispatch()
  const [showFrequencyPlanWarning, setShowFrequencyPlanWarning] = React.useState(
    isEmptyFrequencyPlan(frequency_plan_ids) || !frequency_plan_ids,
  )

  useEffect(() => {
    dispatch(getGsFrequencyPlans())
  }, [dispatch])

  const freqPlanOptions = [
    ...formatOptions(useSelector(selectGsFrequencyPlans)),
    { value: 'no-frequency-plan', label: m.none },
  ]

  const onFrequencyPlanChange = useCallback(freqPlan => {
    setShowFrequencyPlanWarning(isEmptyFrequencyPlan(freqPlan))
  }, [])

  return (
    <Form.Field
      name="frequency_plan_ids"
      title={sharedMessages.frequencyPlan}
      description={m.frequencyPlanDescription}
      valuePlaceholder={m.selectFrequencyPlan}
      tooltipId={tooltipIds.FREQUENCY_PLAN}
      warning={showFrequencyPlanWarning ? sharedMessages.frequencyPlanWarning : undefined}
      component={KeyValueMap}
      inputElement={Select}
      indexAsKey
      addMessage={m.addFrequencyPlan}
      onChange={onFrequencyPlanChange}
      additionalInputProps={{ options: freqPlanOptions }}
      distinctOptions
      atLeastOneEntry
      filterByTag
      required
    />
  )
}

export default GatewayFrequencyPlansSelect
