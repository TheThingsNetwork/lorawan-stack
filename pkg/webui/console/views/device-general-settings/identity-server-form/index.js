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

import SubmitButton from '../../../../components/submit-button'
import SubmitBar from '../../../../components/submit-bar'
import Input from '../../../../components/input'
import Form from '../../../../components/form'
import Checkbox from '../../../../components/checkbox'

import diff from '../../../../lib/diff'
import m from '../../../components/device-data-form/messages'
import PropTypes from '../../../../lib/prop-types'
import sharedMessages from '../../../../lib/shared-messages'

import { parseLorawanMacVersion, hasExternalJs, isDeviceOTAA } from '../utils'
import validationSchema from './validation-schema'

const IdentityServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess, jsConfig } = props

  const formRef = React.useRef(null)
  const [error, setError] = React.useState('')
  const [externalJs, setExternaljs] = React.useState(hasExternalJs(device))

  const initialValues = React.useMemo(() => {
    const extJs = hasExternalJs(device)
    const {
      ids,
      name,
      description,
      network_server_address,
      application_server_address,
      join_server_address,
      lorawan_version,
      // JS form fields that should be reset when provisioning devices on an external JS.
      resets_join_nonces,
      root_keys = {
        nwk_key: {},
        app_key: {},
      },
    } = device

    return {
      name,
      description,
      application_server_address,
      network_server_address,
      _lorawan_version: lorawan_version,
      join_server_address: extJs ? undefined : join_server_address,
      _external_js: extJs,
      ids,
      // JS form fields that should be reset when provisioning devices on an external JS.
      root_keys,
      resets_join_nonces,
    }
  }, [device])

  const handleExternalJsChange = React.useCallback(
    evt => {
      const { checked: externalJsChecked } = evt.target
      const { setValues, state: formState } = formRef.current

      setExternaljs(externalJsChecked)

      // Note: If the end device is provisioned on an external JS, we reset `root_keys` and
      // `resets_join_nonces` fields.
      if (externalJsChecked) {
        setValues({
          ...formState.values,
          root_keys: {
            nwk_key: {},
            app_key: {},
          },
          resets_join_nonces: false,
          join_server_address: '',
          _external_js: externalJsChecked,
        })
      } else {
        let { join_server_address } = initialValues
        const { resets_join_nonces, root_keys } = initialValues
        // if JS address is not set, always fallback to the default js address
        // when resetting from the 'provisioned by external js' option.
        if (!Boolean(join_server_address)) {
          join_server_address = new URL(jsConfig.base_url).hostname || ''
        }

        setValues({
          ...formState.values,
          join_server_address,
          root_keys,
          _external_js: externalJsChecked,
          resets_join_nonces,
        })
      }
    },
    [initialValues, jsConfig.base_url],
  )

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, ['_external_js', '_lorawan_version'])

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
      </SubmitBar>
    </Form>
  )
})

IdentityServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default IdentityServerForm
