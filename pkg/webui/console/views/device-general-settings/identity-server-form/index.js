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
import * as Yup from 'yup'

import SubmitButton from '../../../../components/submit-button'
import SubmitBar from '../../../../components/submit-bar'
import Input from '../../../../components/input'
import Form from '../../../../components/form'

import diff from '../../../../lib/diff'
import m from '../../../components/device-data-form/messages'
import { id as deviceIdRegexp, address as addressRegexp } from '../../../lib/regexp'
import PropTypes from '../../../../lib/prop-types'
import sharedMessages from '../../../../lib/shared-messages'

import { parseLorawanMacVersion, hasExternalJs, isDeviceOTAA } from '../utils'

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    device_id: Yup.string()
      .matches(deviceIdRegexp, sharedMessages.validateAlphanum)
      .min(2, sharedMessages.validateTooShort)
      .max(36, sharedMessages.validateTooLong)
      .required(sharedMessages.validateRequired),
  }),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string().max(2000, sharedMessages.validateTooLong),
  network_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
  application_server_address: Yup.string().matches(
    addressRegexp,
    sharedMessages.validateAddressFormat,
  ),
  join_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
})

const IdentityServerForm = React.memo(props => {
  const { device, onSubmit } = props

  const [error, setError] = React.useState('')

  const initialValues = React.useMemo(() => {
    const extJs = hasExternalJs(device)
    const {
      ids,
      name,
      description,
      network_server_address,
      application_server_address,
      join_server_address,
    } = device

    return {
      name,
      description,
      application_server_address,
      network_server_address,
      join_server_address: extJs ? undefined : join_server_address,
      ids,
    }
  }, [device])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues)

      setError('')
      try {
        await onSubmit(updatedValues)
        resetForm(castedValues)
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit],
  )

  const isOTAA = isDeviceOTAA(device)
  const externalJs = hasExternalJs(device)
  const isNewLorawanVersion = parseLorawanMacVersion(device.lorawan_version) >= 110
  const hasJoinEUI = Boolean(device.ids.join_eui)
  const hasDevEUI = Boolean(device.ids.dev_eui)

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      enableReinitialize
    >
      <Form.Field
        title={sharedMessages.devID}
        name="ids.device_id"
        placeholder={m.deviceIdPlaceholder}
        description={m.deviceIdDescription}
        required
        disabled
        component={Input}
      />
      {hasJoinEUI && (
        <Form.Field
          title={isNewLorawanVersion ? sharedMessages.joinEUI : sharedMessages.appEUI}
          name="ids.join_eui"
          type="byte"
          min={8}
          max={8}
          description={isNewLorawanVersion ? m.joinEUIDescription : m.appEUIDescription}
          required
          disabled
          component={Input}
        />
      )}
      {hasDevEUI && (
        <Form.Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          description={m.deviceEUIDescription}
          required
          disabled
          component={Input}
        />
      )}
      <Form.Field
        title={sharedMessages.devName}
        name="name"
        placeholder={m.deviceNamePlaceholder}
        description={m.deviceNameDescription}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.devDesc}
        name="description"
        type="textarea"
        description={m.deviceDescDescription}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.applicationServerAddress}
        placeholder={sharedMessages.addressPlaceholder}
        name="application_server_address"
        component={Input}
      />
      <Form.Field
        title={sharedMessages.networkServerAddress}
        placeholder={sharedMessages.addressPlaceholder}
        name="network_server_address"
        component={Input}
      />
      <Form.Field
        title={sharedMessages.joinServerAddress}
        placeholder={isOTAA && externalJs ? m.external : sharedMessages.addressPlaceholder}
        name="join_server_address"
        component={Input}
        disabled={!isOTAA || externalJs}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

IdentityServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  onSubmit: PropTypes.func.isRequired,
}

export default IdentityServerForm
