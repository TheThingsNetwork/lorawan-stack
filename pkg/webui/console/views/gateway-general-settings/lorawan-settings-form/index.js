// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { GsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { unit as unitRegexp, emptyDuration as emptyDurationRegexp } from '@console/lib/regexp'

import validationSchema from './validation-schema'

const decodeDelayValue = value => {
  if (emptyDurationRegexp.test(value)) {
    return {
      duration: undefined,
      unit: value,
    }
  }
  const duration = value.split(unitRegexp)[0]
  const unit = value.split(duration)[1]
  return {
    duration: duration ? Number(duration) : undefined,
    unit,
  }
}

const isEmptyFrequencyPlan = value => value === frequencyPlans.EMPTY_FREQ_PLAN

const isNotValidDuration = value => {
  const { duration, unit } = decodeDelayValue(value)

  switch (unit) {
    case 'ms':
      return duration < delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY
    case 's':
      return duration < delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY / 1000
    case 'm':
      return duration < delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY / 60000
    case 'h':
      return duration < delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY / 3600000
  }
}

const LorawanSettingsForm = React.memo(props => {
  const { gateway, onSubmit, onSubmitSuccess } = props

  const [error, setError] = React.useState(undefined)

  const [shouldDisplayWarning, setShouldDisplayWarning] = React.useState(
    isNotValidDuration(gateway.schedule_anytime_delay),
  )

  const onScheduleAnytimeDelayChange = React.useCallback(value => {
    setShouldDisplayWarning(isNotValidDuration(value))
  }, [])

  const [showFrequencyPlanWarning, setShowFrequencyPlanWarning] = React.useState(
    isEmptyFrequencyPlan(gateway.frequency_plan_id) || !gateway.frequency_plan_id,
  )

  const onFrequencyPlanChange = React.useCallback(freqPlan => {
    setShowFrequencyPlanWarning(isEmptyFrequencyPlan(freqPlan.value))
  }, [])

  const initialValues = React.useMemo(() => {
    return {
      ...validationSchema.cast(gateway),
      frequency_plan_id: gateway.frequency_plan_id || frequencyPlans.EMPTY_FREQ_PLAN,
    }
  }, [gateway])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(
        isEmptyFrequencyPlan(values.frequency_plan_id)
          ? { ...values, frequency_plan_id: '' }
          : values,
      )

      setError(undefined)
      try {
        await onSubmit(castedValues)
        resetForm({ values: castedValues })
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [onSubmit, onSubmitSuccess],
  )

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      enableReinitialize
    >
      <GsFrequencyPlansSelect
        name="frequency_plan_id"
        menuPlacement="top"
        onChange={onFrequencyPlanChange}
        warning={showFrequencyPlanWarning ? sharedMessages.frequencyPlanWarning : undefined}
        glossaryId={glossaryIds.FREQUENCY_PLAN}
      />
      <Form.Field
        title={sharedMessages.gatewayScheduleDownlinkLate}
        name="schedule_downlink_late"
        component={Checkbox}
        description={sharedMessages.scheduleDownlinkLateDescription}
      />
      <Form.Field
        title={sharedMessages.dutyCycle}
        name="enforce_duty_cycle"
        component={Checkbox}
        label={sharedMessages.enforced}
        description={sharedMessages.enforceDutyCycleDescription}
      />
      <Form.Field
        title={sharedMessages.scheduleAnyTimeDelay}
        name="schedule_anytime_delay"
        inputWidth="s"
        component={UnitInput}
        description={{
          ...sharedMessages.scheduleAnyTimeDescription,
          values: {
            minimumValue: delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY,
            defaultValue: delay.DEFAULT_GATEWAY_SCHEDULE_ANYTIME_DELAY,
          },
        }}
        units={[
          { label: sharedMessages.milliseconds, value: 'ms' },
          { label: sharedMessages.seconds, value: 's' },
          { label: sharedMessages.minutes, value: 'm' },
          { label: sharedMessages.hours, value: 'h' },
        ]}
        onChange={onScheduleAnytimeDelayChange}
        warning={
          shouldDisplayWarning
            ? {
                ...sharedMessages.delayWarning,
                values: { minimumValue: delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY },
              }
            : undefined
        }
        required
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
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default LorawanSettingsForm
