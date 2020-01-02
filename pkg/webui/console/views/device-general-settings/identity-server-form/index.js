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
import { defineMessages } from 'react-intl'

import SubmitButton from '../../../../components/submit-button'
import SubmitBar from '../../../../components/submit-bar'
import Input from '../../../../components/input'
import Form from '../../../../components/form'
import Checkbox from '../../../../components/checkbox'
import ModalButton from '../../../../components/button/modal-button'

import diff from '../../../../lib/diff'
import m from '../../../components/device-data-form/messages'
import PropTypes from '../../../../lib/prop-types'
import sharedMessages from '../../../../lib/shared-messages'
import { parseLorawanMacVersion, hasExternalJs, isDeviceOTAA } from '../utils'
import validationSchema from './validation-schema'

const messages = defineMessages({
  deleteDevice: 'Delete End Device',
  deleteWarning:
    'Are you sure you want to delete "{deviceId}"? This action cannot be undone and it will not be possible to reuse the end device ID!',
})

const IdentityServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess, onDelete, onDeleteSuccess, onDeleteFailure } = props
  const { name, ids } = device

  const formRef = React.useRef(null)
  const [error, setError] = React.useState('')
  const [externalJs, setExternaljs] = React.useState(hasExternalJs(device))

  const initialValues = React.useMemo(() => {
    const initialValues = {
      ...device,
      _external_js: hasExternalJs(device),
      _lorawan_version: device.lorawan_version,
      _supports_join: device.supports_join,
    }

    return validationSchema.cast(initialValues)
  }, [device])

  const handleExternalJsChange = React.useCallback(evt => {
    const { checked: externalJsChecked } = evt.target
    const { setValues, state } = formRef.current

    setExternaljs(externalJsChecked)

    const values = {
      ...state.values,
      _external_js: externalJsChecked,
    }

    setValues(validationSchema.cast(values))
  }, [])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, [
        '_external_js',
        '_lorawan_version',
        '_supports_join',
      ])

      setError('')
      try {
        await onSubmit(updatedValues)
        resetForm(castedValues)
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit, onSubmitSuccess],
  )

  const onDeviceDelete = React.useCallback(async () => {
    try {
      await onDelete()
      onDeleteSuccess()
    } catch (error) {
      onDeleteFailure()
    }
  }, [onDelete, onDeleteFailure, onDeleteSuccess])

  const isOTAA = isDeviceOTAA(device)
  const isNewLorawanVersion = parseLorawanMacVersion(device.lorawan_version) >= 110
  const hasJoinEUI = Boolean(device.ids.join_eui)
  const hasDevEUI = Boolean(device.ids.dev_eui)

  let joinServerAddressPlaceholder = sharedMessages.addressPlaceholder
  if (isOTAA && externalJs) {
    joinServerAddressPlaceholder = m.external
  } else if (!isOTAA) {
    joinServerAddressPlaceholder = sharedMessages.empty
  }

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      formikRef={formRef}
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
        title={sharedMessages.networkServerAddress}
        placeholder={sharedMessages.addressPlaceholder}
        name="network_server_address"
        component={Input}
      />
      <Form.Field
        title={sharedMessages.applicationServerAddress}
        placeholder={sharedMessages.addressPlaceholder}
        name="application_server_address"
        component={Input}
      />
      {isOTAA && (
        <>
          <Form.Field
            title={m.externalJoinServer}
            description={m.externalJoinServerDescription}
            name="_external_js"
            onChange={handleExternalJsChange}
            component={Checkbox}
          />
          <Form.Field
            title={sharedMessages.joinServerAddress}
            placeholder={joinServerAddressPlaceholder}
            name="join_server_address"
            component={Input}
            disabled={!isOTAA || externalJs}
          />
        </>
      )}
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
        <ModalButton
          type="button"
          icon="delete"
          message={messages.deleteDevice}
          modalData={{
            message: { values: { deviceId: name || ids.device_id }, ...messages.deleteWarning },
          }}
          onApprove={onDeviceDelete}
          danger
        />
      </SubmitBar>
    </Form>
  )
})

IdentityServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  onDelete: PropTypes.func.isRequired,
  onDeleteFailure: PropTypes.func.isRequired,
  onDeleteSuccess: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default IdentityServerForm
