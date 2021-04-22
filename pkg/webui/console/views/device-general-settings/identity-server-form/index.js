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

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'

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
    jsConfig,
    nsConfig,
    asConfig,
  } = props
  const { name, ids } = device

  const formRef = React.useRef(null)
  // Store default join server address that is used to fill the `join_server_address` field when
  // `_external_js` checkbox is unchecked.
  const jsAddressRef = React.useRef(
    device.join_server_address
      ? device.join_server_address
      : jsConfig.enabled
      ? getHostnameFromUrl(jsConfig.base_url)
      : '',
  )

  const [error, setError] = React.useState('')
  const [externalJs, setExternaljs] = React.useState(hasExternalJs(device))

  const validationContext = React.useMemo(
    () => ({
      lorawanVersion: device.lorawan_version,
      supportsJoin: device.supports_join,
    }),
    [device.lorawan_version, device.supports_join],
  )

  const initialValues = React.useMemo(() => {
    const initialValues = {
      ...device,
      _external_js: hasExternalJs(device),
      attributes: mapAttributesToFormValue(device.attributes),
    }

    return validationSchema.cast(initialValues, { context: validationContext })
  }, [device, validationContext])

  const handleExternalJsChange = React.useCallback(
    evt => {
      const { checked: externalJsChecked } = evt.target
      const { setValues, values } = formRef.current

      setExternaljs(externalJsChecked)
      setValues(
        validationSchema.cast(
          {
            ...values,
            _external_js: externalJsChecked,
            join_server_address: externalJsChecked ? undefined : jsAddressRef.current,
          },
          { context: validationContext },
        ),
      )
    },
    [validationContext],
  )

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values, { context: validationContext })
      const updatedValues = diff(initialValues, castedValues, ['_external_js'])

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
    [initialValues, onSubmit, onSubmitSuccess, validationContext],
  )

  const onDeviceDelete = React.useCallback(async () => {
    try {
      await onDelete()
      onDeleteSuccess()
    } catch (error) {
      onDeleteFailure()
    }
  }, [onDelete, onDeleteFailure, onDeleteSuccess])

  const { enabled: jsEnabled } = jsConfig
  const { enabled: asEnabled } = asConfig
  const { enabled: nsEnabled } = nsConfig

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
  if (lorawanVersion >= 100 && lorawanVersion < 104) {
    joinEUITitle = sharedMessages.appEUI
  } else if (lorawanVersion >= 104) {
    joinEUITitle = sharedMessages.joinEUI
  }

  return (
    <Form
      validationSchema={validationSchema}
      validationContext={validationContext}
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
          required
          disabled
          component={Input}
          glossaryId={glossaryIds.JOIN_EUI}
        />
      )}
      {hasDevEUI && (
        <Form.Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          required
          disabled
          component={Input}
          glossaryId={glossaryIds.DEV_EUI}
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
        autoComplete="on"
      />
      <Form.Field
        title={sharedMessages.applicationServerAddress}
        placeholder={sharedMessages.addressPlaceholder}
        name="application_server_address"
        component={Input}
        autoComplete="on"
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
            autoComplete="on"
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
  asConfig: PropTypes.stackComponent.isRequired,
  device: PropTypes.device.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  nsConfig: PropTypes.stackComponent.isRequired,
  onDelete: PropTypes.func.isRequired,
  onDeleteFailure: PropTypes.func.isRequired,
  onDeleteSuccess: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default IdentityServerForm
