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
import Field from '../../../components/field'
import SubmitBar from '../../../components/submit-bar'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import FrequencyPlansSelect from '../../containers/freq-plans-select'
import sharedMessages from '../../../lib/shared-messages'
import { id as gatewayIdRegexp, address as addressRegexp } from '../../lib/regexp'

const m = defineMessages({
  dutyCycle: 'Enforce Duty Cycle',
  gatewayIdPlaceholder: 'my-new-gateway',
  gsServerAddressDescription: 'The address of the Gateway Server to connect to',
})

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    gateway_id: Yup.string()
      .matches(gatewayIdRegexp, sharedMessages.validateAlphanum)
      .min(2, sharedMessages.validateTooShort)
      .max(36, sharedMessages.validateTooLong)
      .required(sharedMessages.validateRequired),
    eui: Yup.string()
      .length(8 * 2, sharedMessages.validateTooShort),
  }),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string()
    .max(2000, sharedMessages.validateTooLong),
  frequency_plan_id: Yup.string()
    .required(sharedMessages.validateRequired),
  gateway_server_address: Yup.string()
    .matches(addressRegexp, sharedMessages.validateAddressFormat),
})

@bind
class GatewayDataForm extends React.Component {
  render () {
    const {
      onSubmit,
      update,
      error,
      initialValues,
      mapErrorsToFields,
      formRef,
      children,
    } = this.props

    return (
      <Form
        submitEnabledWhenInvalid
        error={error}
        onSubmit={onSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
        mapErrorsToFields={mapErrorsToFields}
        horizontal
        formikRef={formRef}
      >
        <Message
          component="h4"
          content={sharedMessages.generalSettings}
        />
        <Field
          title={sharedMessages.gatewayID}
          name="ids.gateway_id"
          type="text"
          placeholder={m.gatewayIdPlaceholder}
          required
          disabled={update}
          autoFocus={!update}
        />
        <Field
          title={sharedMessages.gatewayEUI}
          name="ids.eui"
          type="byte"
          min={8}
          max={8}
          placeholder={sharedMessages.gatewayEUI}
          autoFocus={update}
        />
        <Field
          title={sharedMessages.gatewayName}
          name="name"
        />
        <Field
          title={sharedMessages.gatewayDescription}
          name="description"
          type="textarea"
        />
        <Field
          title={sharedMessages.gatewayServerAddress}
          description={m.gsServerAddressDescription}
          placeholder={sharedMessages.addressPlaceholder}
          name="gateway_server_address"
          type="text"
        />
        <Message
          component="h4"
          content={sharedMessages.lorawanOptions}
        />
        <FrequencyPlansSelect
          horizontal
          source="gs"
          name="frequency_plan_id"
          menuPlacement="top"
          required
        />
        <Field
          title={m.dutyCycle}
          name="enforce_duty_cycle"
          type="checkbox"
        />
        <SubmitBar>
          {children}
        </SubmitBar>
      </Form>
    )
  }
}

GatewayDataForm.propTypes = {
  update: PropTypes.bool,
  error: PropTypes.error,
  initialValues: PropTypes.object,
  mapErrorsToFields: PropTypes.object,
  onSubmit: PropTypes.func.isRequired,
  /** React reference to be passed to the form */
  formRef: PropTypes.object,
  /** SubmitBar contents */
  children: PropTypes.node.isRequired,
}

GatewayDataForm.defaultProps = {
  update: false,
  error: '',
  initialValues: {},
  mapErrorsToFields: {},
}

export default GatewayDataForm
