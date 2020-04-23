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

import React from 'react'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import * as Yup from 'yup'

import delay from '@console/constants/delays'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import UnitInput from '@ttn-lw/components/unit-input'

import Message from '@ttn-lw/lib/components/message'

import { GsFrequencyPlansSelect } from '@console/containers/freq-plans-select'
import OwnersSelect from '@console/containers/owners-select'

import {
  id as gatewayIdRegexp,
  address as addressRegexp,
  unit as unitRegexp,
  emptyDuration as emptyDurationRegexp,
  delay as delayRegexp,
} from '@console/lib/regexp'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  enforced: 'Enforced',
  dutyCycle: 'Duty cycle',
  gatewayIdPlaceholder: 'my-new-gateway',
  gatewayNamePlaceholder: 'My new gateway',
  gsServerAddressDescription: 'The address of the Gateway Server to connect to',
  gatewayDescPlaceholder: 'Description for my new gateway',
  gatewayDescDescription:
    'Optional gateway description; can also be used to save notes about the gateway',
  statusDescription: 'The status of this gateway may be publicly displayed',
  scheduleDownlinkLateDescription: 'Enable server-side buffer of downlink messages',
  autoUpdateDescription: 'Gateway can be updated automatically',
  updateChannelDescription: 'Channel for gateway automatic updates',
  enforceDutyCycleDescription:
    'Recommended for all gateways in order to respect spectrum regulations',
  scheduleAnyTimeDelay: 'Schedule any time delay',
  scheduleAnyTimeDescription:
    'Configure gateway delay (minimum: {minimumValue}ms, default: {defaultValue}ms)',
  miliseconds: 'miliseconds',
  seconds: 'seconds',
  minutes: 'minutes',
  hours: 'hours',
  delayWarning:
    'Delay too short. The lower bound ({minimumValue}ms) will be used by the Gateway Server.',
})

const validationSchema = Yup.object().shape({
  owner_id: Yup.string(),
  ids: Yup.object().shape({
    gateway_id: Yup.string()
      .matches(gatewayIdRegexp, sharedMessages.validateIdFormat)
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .required(sharedMessages.validateRequired),
    eui: Yup.nullableString().length(8 * 2, Yup.passValues(sharedMessages.validateLength)),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  update_channel: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  gateway_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
  location_public: Yup.boolean().default(false),
  status_public: Yup.boolean().default(false),
  schedule_downlink_late: Yup.boolean().default(false),
  update_location_from_status: Yup.boolean().default(false),
  auto_update: Yup.boolean().default(false),
  schedule_anytime_delay: Yup.string().matches(delayRegexp, sharedMessages.validateDelayFormat),
})

class GatewayDataForm extends React.Component {
  static propTypes = {
    /** The SubmitBar content. */
    children: PropTypes.node.isRequired,
    error: PropTypes.error,
    /** React reference to be passed to the form. */
    formRef: PropTypes.shape({}),
    initialValues: PropTypes.gateway,
    onSubmit: PropTypes.func.isRequired,
    update: PropTypes.bool,
  }

  static defaultProps = {
    formRef: undefined,
    update: false,
    error: '',
    initialValues: validationSchema.cast({}),
  }

  constructor(props) {
    super(props)

    this.state = {
      shouldDisplayWarning: this.isNotValidDuration(props.initialValues.schedule_anytime_delay),
    }
  }

  @bind
  onScheduleAnytimeDelayChange(value) {
    this.setState({ shouldDisplayWarning: this.isNotValidDuration(value) })
  }

  @bind
  onSubmit(values, helpers) {
    const { onSubmit } = this.props
    const castedValues = validationSchema.cast(values)

    onSubmit(castedValues, helpers)
  }

  decodeDelayValue(value) {
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

  isNotValidDuration(value) {
    const { duration, unit } = this.decodeDelayValue(value)
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

  render() {
    const { update, error, initialValues, formRef, children } = this.props
    const { shouldDisplayWarning } = this.state

    return (
      <Form
        error={error}
        onSubmit={this.onSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
        formikRef={formRef}
      >
        <Message component="h4" content={sharedMessages.generalSettings} />
        {!update && <OwnersSelect name="owner_id" required autoFocus />}
        <Form.Field
          title={sharedMessages.gatewayID}
          name="ids.gateway_id"
          placeholder={m.gatewayIdPlaceholder}
          required
          disabled={update}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.gatewayEUI}
          name="ids.eui"
          type="byte"
          min={8}
          max={8}
          placeholder={sharedMessages.gatewayEUI}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.gatewayName}
          placeholder={m.gatewayNamePlaceholder}
          name="name"
          component={Input}
        />
        <Form.Field
          title={sharedMessages.gatewayDescription}
          description={m.gatewayDescDescription}
          placeholder={m.gatewayDescPlaceholder}
          name="description"
          type="textarea"
          component={Input}
        />
        <Form.Field
          title={sharedMessages.gatewayServerAddress}
          description={m.gsServerAddressDescription}
          placeholder={sharedMessages.addressPlaceholder}
          name="gateway_server_address"
          component={Input}
        />
        <Form.Field
          title={sharedMessages.gatewayStatus}
          name="status_public"
          component={Checkbox}
          label={sharedMessages.public}
          description={m.statusDescription}
        />
        <Message component="h4" content={sharedMessages.lorawanOptions} />
        <GsFrequencyPlansSelect name="frequency_plan_id" menuPlacement="top" required />
        <Form.Field
          title={sharedMessages.gatewayScheduleDownlinkLate}
          name="schedule_downlink_late"
          component={Checkbox}
          description={m.scheduleDownlinkLateDescription}
        />
        <Form.Field
          title={m.dutyCycle}
          name="enforce_duty_cycle"
          component={Checkbox}
          label={m.enforced}
          description={m.enforceDutyCycleDescription}
        />
        <Form.Field
          title={m.scheduleAnyTimeDelay}
          name="schedule_anytime_delay"
          component={UnitInput}
          description={{
            ...m.scheduleAnyTimeDescription,
            values: {
              minimumValue: delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY,
              defaultValue: delay.DEFAULT_GATEWAY_SCHEDULE_ANYTIME_DELAY,
            },
          }}
          units={[
            { label: m.miliseconds, value: 'ms' },
            { label: m.seconds, value: 's' },
            { label: m.minutes, value: 'm' },
            { label: m.hours, value: 'h' },
          ]}
          onChange={this.onScheduleAnytimeDelayChange}
          decode={this.decodeDelayValue}
          warning={
            shouldDisplayWarning
              ? {
                  ...m.delayWarning,
                  values: { minimumValue: delay.MINIMUM_GATEWAY_SCHEDULE_ANYTIME_DELAY },
                }
              : undefined
          }
          required
        />
        <Message component="h4" content={sharedMessages.gatewayUpdateOptions} />
        <Form.Field
          title={sharedMessages.automaticUpdates}
          name="auto_update"
          component={Checkbox}
          description={m.autoUpdateDescription}
        />
        <Form.Field
          title={sharedMessages.channel}
          description={m.updateChannelDescription}
          placeholder={sharedMessages.stable}
          name="update_channel"
          component={Input}
        />
        <SubmitBar>{children}</SubmitBar>
      </Form>
    )
  }
}

export default GatewayDataForm
