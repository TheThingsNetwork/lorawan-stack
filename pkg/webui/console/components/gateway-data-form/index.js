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

import Form from '../../../components/form'
import Input from '../../../components/input'
import Checkbox from '../../../components/checkbox'
import SubmitBar from '../../../components/submit-bar'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import { GsFrequencyPlansSelect } from '../../containers/freq-plans-select'
import sharedMessages from '../../../lib/shared-messages'
import { id as gatewayIdRegexp, address as addressRegexp } from '../../lib/regexp'
import OwnersSelect from '../../containers/owners-select'

const m = defineMessages({
  enforced: 'Enforced',
  dutyCycle: 'Duty Cycle',
  gatewayIdPlaceholder: 'my-new-gateway',
  gatewayNamePlaceholder: 'My New Gateway',
  gsServerAddressDescription: 'The address of the Gateway Server to connect to',
  gatewayDescPlaceholder: 'Description for my new gateway',
  gatewayDescDescription:
    'Optional gateway description; can also be used to save notes about the gateway',
  statusDescription: 'The status of this gateway may be publicly displayed',
  locationDescription: 'The location of this gateway may be publicly displayed',
  scheduleDownlinkLateDescription: 'Enable server-side buffer of downlink messages',
  autoUpdateDescription: 'Gateway can be updated automatically',
  updateChannelDescription: 'Channel for gateway automatic updates',
  enforceDutyCycleDescription:
    'Recommended for all gateways in order to respect spectrum regulations',
})

const validationSchema = Yup.object().shape({
  owner_id: Yup.string(),
  ids: Yup.object().shape({
    gateway_id: Yup.string()
      .matches(gatewayIdRegexp, sharedMessages.validateIdFormat)
      .min(2, sharedMessages.validateTooShort)
      .max(36, sharedMessages.validateTooLong)
      .required(sharedMessages.validateRequired),
    eui: Yup.nullableString().length(8 * 2, sharedMessages.validateTooShort),
  }),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  update_channel: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string().max(2000, sharedMessages.validateTooLong),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  gateway_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
  location_public: Yup.boolean().default(false),
  status_public: Yup.boolean().default(false),
  schedule_downlink_late: Yup.boolean().default(false),
  auto_update: Yup.boolean().default(false),
})

@bind
class GatewayDataForm extends React.Component {
  onSubmit(values, helpers) {
    const { onSubmit } = this.props
    const castedValues = validationSchema.cast(values)

    onSubmit(castedValues, helpers)
  }

  render() {
    const { update, error, initialValues, formRef, children } = this.props

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
          title={sharedMessages.gatewayLocation}
          name="location_public"
          component={Checkbox}
          label={sharedMessages.public}
          description={m.locationDescription}
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

GatewayDataForm.propTypes = {
  children: PropTypes.node.isRequired,
  error: PropTypes.error,
  formRef: PropTypes.shape({}),
  initialValues: PropTypes.shape({}),
  /** React reference to be passed to the form */
  onSubmit: PropTypes.func.isRequired,
  /** SubmitBar contents */
  update: PropTypes.bool,
}

GatewayDataForm.defaultProps = {
  formRef: undefined,
  update: false,
  error: '',
  initialValues: {},
}

export default GatewayDataForm
