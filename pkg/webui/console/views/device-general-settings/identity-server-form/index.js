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

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import ModalButton from '@ttn-lw/components/button/modal-button'
import KeyValueMap from '@ttn-lw/components/key-value-map'

import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'

import { mapFormValueToAttributes, mapAttributesToFormValue } from '@console/lib/attributes'
import { parseLorawanMacVersion } from '@console/lib/device-utils'

import { hasExternalJs, isDeviceOTAA } from '../utils'

import validationSchema from './validation-schema'

const messages = defineMessages({
  deleteDevice: 'Delete end device',
  deleteWarning:
    'Are you sure you want to delete "{deviceId}"? This action cannot be undone and it will not be possible to reuse the end device ID.',
})

const IdentityServerForm = React.memo(props => {
  const {
    device,
    onSubmit,
    onSubmitSuccess,
    onDelete,
    onDeleteSuccess,
    onDeleteFailure,
    mayReadKeys,
  } = props
  const { name, ids } = device

  const formRef = React.useRef(null)
  const [error, setError] = React.useState('')
  const [externalJs, setExternaljs] = React.useState(hasExternalJs(device) && mayReadKeys)

  const initialValues = React.useMemo(() => {
    const initialValues = {
      ...device,
      _external_js: hasExternalJs(device) && mayReadKeys,
      _lorawan_version: device.lorawan_version,
      _supports_join: device.supports_join,
      attributes: mapAttributesToFormValue(device.attributes),
    }

    return validationSchema.cast(initialValues)
  }, [device, mayReadKeys])

  const handleExternalJsChange = React.useCallback(evt => {
    const { checked: externalJsChecked } = evt.target
    const { setValues, values } = formRef.current

    setExternaljs(externalJsChecked)

    setValues(validationSchema.cast({ ...values, _external_js: externalJsChecked }))
  }, [])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, [
        '_external_js',
        '_lorawan_version',
        '_supports_join',
      ])

      const update =
        'attributes' in updatedValues
          ? { ...updatedValues, attributes: mapFormValueToAttributes(values.attributes) }
          : updatedValues

      setError('')
      try {
        await onSubmit(update)
        resetForm({ values: castedValues })
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

  const { enabled: jsEnabled } = selectJsConfig()
  const { enabled: asEnabled } = selectAsConfig()
  const { enabled: nsEnabled } = selectNsConfig()

  const lorawanVersion = parseLorawanMacVersion(device.lorawan_version)
  const isOTAA = isDeviceOTAA(device)
  const hasJoinEUI = Boolean(device.ids.join_eui)
  const hasDevEUI = Boolean(device.ids.dev_eui)

  // We do not want to show the external JS option if the user is on JS only
  // deployment.
  // See https://github.com/TheThingsNetwork/lorawan-stack/issues/2119#issuecomment-597736420
  const hideExternalJs = !isOTAA || (jsEnabled && !asEnabled && !nsEnabled)

  let joinServerAddressPlaceholder = sharedMessages.addressPlaceholder
  if (isOTAA && externalJs) {
    joinServerAddressPlaceholder = sharedMessages.external
  } else if (!isOTAA) {
    joinServerAddressPlaceholder = sharedMessages.empty
  }

  let joinEUITitle = sharedMessages.appEUIJoinEUI
  let joinEUIDescription
  if (lorawanVersion >= 100 && lorawanVersion < 104) {
    joinEUITitle = sharedMessages.appEUI
    joinEUIDescription = sharedMessages.appEUIDescription
  } else if (lorawanVersion >= 104) {
    joinEUITitle = sharedMessages.joinEUI
    joinEUIDescription = sharedMessages.joinEUIDescription
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
        placeholder={sharedMessages.deviceIdPlaceholder}
        description={sharedMessages.deviceIdDescription}
        required
        disabled
        component={Input}
      />
      {hasJoinEUI && (
        <Form.Field
          title={joinEUITitle}
          name="ids.join_eui"
          type="byte"
          min={8}
          max={8}
          description={joinEUIDescription}
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
          description={sharedMessages.deviceEUIDescription}
          required
          disabled
          component={Input}
        />
      )}
      <Form.Field
        title={sharedMessages.devName}
        name="name"
        placeholder={sharedMessages.deviceNamePlaceholder}
        description={sharedMessages.deviceNameDescription}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.devDesc}
        name="description"
        type="textarea"
        description={sharedMessages.deviceDescDescription}
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
      {!hideExternalJs && (
        <>
          <Form.Field
            title={sharedMessages.externalJoinServer}
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
      <Form.Field
        name="attributes"
        title={sharedMessages.attributes}
        keyPlaceholder={sharedMessages.key}
        valuePlaceholder={sharedMessages.value}
        addMessage={sharedMessages.addAttributes}
        component={KeyValueMap}
        description={sharedMessages.attributeDescription}
      />
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
          naked
          danger
        />
      </SubmitBar>
    </Form>
  )
})

IdentityServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  mayReadKeys: PropTypes.bool.isRequired,
  onDelete: PropTypes.func.isRequired,
  onDeleteFailure: PropTypes.func.isRequired,
  onDeleteSuccess: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default IdentityServerForm
