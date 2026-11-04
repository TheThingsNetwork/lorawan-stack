// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'

import delay from '@console/constants/delays'
import frequencyPlans from '@console/constants/frequency-plans'

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import UnitInput from '@ttn-lw/components/unit-input'

import GatewayFrequencyPlansSelect from '@console/containers/freq-plans-select/gs-frequency-plan-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import validationSchema from './validation-schema'

const isEmptyFrequencyPlan = value => value?.includes(frequencyPlans.EMPTY_FREQ_PLAN)

// A zero delay lets the Gateway Server use the default delay, so the field is left empty
// to show the default delay as placeholder instead.
const decodeDelay = value => (parseFloat(value) ? value : '')

const LorawanSettingsForm = React.memo(props => {
  const { gateway, onSubmit } = props

  const [error, setError] = React.useState(undefined)

  const initialValues = React.useMemo(
    () => ({
      ...validationSchema.cast(gateway),
      frequency_plan_ids: gateway.frequency_plan_ids || [frequencyPlans.EMPTY_FREQ_PLAN],
      schedule_anytime_delay: decodeDelay(gateway.schedule_anytime_delay),
    }),
    [gateway],
  )

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(
        isEmptyFrequencyPlan(values.frequency_plan_ids)
          ? { ...values, frequency_plan_ids: [''] }
          : values,
      )

      setError(undefined)
      try {
        await onSubmit(castedValues)
        resetForm({
          values: {
            ...castedValues,
            schedule_anytime_delay: decodeDelay(castedValues.schedule_anytime_delay),
          },
        })
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [onSubmit],
  )

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      enableReinitialize
    >
      <GatewayFrequencyPlansSelect />
      <Form.Field
        title={sharedMessages.gatewayScheduleDownlinkLate}
        name="schedule_downlink_late"
        component={Checkbox}
        description={sharedMessages.scheduleDownlinkLateDescription}
        tooltipId={tooltipIds.SCHEDULE_DOWNLINK_LATE}
      />
      <Form.Field
        name="enforce_duty_cycle"
        component={Checkbox}
        label={sharedMessages.enforceDutyCycle}
        description={sharedMessages.enforceDutyCycleDescription}
        tooltipId={tooltipIds.ENFORCE_DUTY_CYCLE}
      />
      <Form.Field
        title={sharedMessages.scheduleAnyTimeDelay}
        name="schedule_anytime_delay"
        component={UnitInput.Duration}
        unitSelector={['ms', 's']}
        description={{
          ...sharedMessages.scheduleAnyTimeDescription,
          values: { minimumValue: delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY },
        }}
        placeholder={`${delay.DEFAULT_GATEWAY_SCHEDULE_ANYTIME_DELAY}ms`}
        units={[
          { label: sharedMessages.milliseconds, value: 'ms' },
          { label: sharedMessages.seconds, value: 's' },
          { label: sharedMessages.minutes, value: 'm' },
          { label: sharedMessages.hours, value: 'h' },
        ]}
        tooltipId={tooltipIds.SCHEDULE_ANYTIME_DELAY}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

LorawanSettingsForm.propTypes = {
  gateway: PropTypes.gateway.isRequired,
  onSubmit: PropTypes.func.isRequired,
}

export default LorawanSettingsForm
