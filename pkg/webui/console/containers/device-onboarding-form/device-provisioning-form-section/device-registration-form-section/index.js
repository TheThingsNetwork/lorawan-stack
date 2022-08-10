// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'

import Input from '@ttn-lw/components/input'
import Form, { useFormContext } from '@ttn-lw/components/form'

import DevEUIComponent from '@console/components/dev-eui-component'

import DevAddrInput from '@console/containers/dev-addr-input'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'
import { checkFromState, mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import messages from '../../messages'

import { devEUISchema, initialValues } from './validation-schema'

const DeviceRegistrationFormSection = () => {
  const { values, setFieldValue } = useFormContext()

  const mayEditKeys = useSelector(state => checkFromState(mayEditApplicationDeviceKeys, state))

  const template = useSelector(selectDeviceTemplate)
  const idInputRef = React.useRef(null)

  const generateDeviceId = (device = {}) => {
    const { ids: idsValues = {} } = device

    try {
      devEUISchema.validateSync(idsValues.dev_eui)
      return `eui-${idsValues.dev_eui.toLowerCase()}`
    } catch (e) {
      // We dont want to use invalid `dev_eui` as `device_id`.
    }

    return initialValues.ids.device_id || ''
  }

  const handleIdTextSelect = React.useCallback(
    idInputRef => {
      if (idInputRef.current && values) {
        const { setSelectionRange } = idInputRef.current

        const generatedId = generateDeviceId(values)
        if (generatedId.length > 0 && generatedId === values.ids.device_id) {
          setSelectionRange.call(idInputRef.current, 0, generatedId.length)
        }
      }
    },
    [values],
  )

  const { end_device } = template
  const { supports_join, lorawan_version } = end_device

  const isOTAA = supports_join
  const lwVersion = parseLorawanMacVersion(lorawan_version)

  let appKeyPlaceholder = undefined
  let nwkKeyPlaceholder = undefined
  if (!mayEditKeys) {
    appKeyPlaceholder = sharedMessages.insufficientAppKeyRights
    nwkKeyPlaceholder = sharedMessages.insufficientNwkKeyRights
  }

  return (
    <div data-test-id="device-registration">
      {isOTAA && (
        <>
          <DevEUIComponent
            values={values}
            setFieldValue={setFieldValue}
            initialValues={initialValues}
            devEUISchema={devEUISchema}
          />
          <Form.Field
            required
            title={sharedMessages.appKey}
            name="root_keys.app_key.key"
            type="byte"
            min={16}
            max={16}
            component={Input.Generate}
            disabled={!mayEditKeys}
            mayGenerateValue={mayEditKeys}
            onGenerateValue={generate16BytesKey}
            tooltipId={tooltipIds.APP_KEY}
            placeholder={appKeyPlaceholder}
          />
          {lwVersion >= 110 && (
            <Form.Field
              required
              title={sharedMessages.nwkKey}
              name="root_keys.nwk_key.key"
              type="byte"
              min={16}
              max={16}
              component={Input.Generate}
              disabled={!mayEditKeys}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
              placeholder={nwkKeyPlaceholder}
              tooltipId={tooltipIds.NETWORK_KEY}
            />
          )}
        </>
      )}
      {!isOTAA && (
        <>
          <DevAddrInput title={sharedMessages.devAddr} name="session.dev_addr" required />
          {lwVersion === 104 && (
            <DevEUIComponent
              values={values}
              setFieldValue={setFieldValue}
              initialValues={initialValues}
              devEUISchema={devEUISchema}
            />
          )}
          <Form.Field
            required={mayEditKeys}
            title={sharedMessages.appSKey}
            name="session.keys.app_s_key.key"
            type="byte"
            min={16}
            max={16}
            component={Input.Generate}
            mayGenerateValue={mayEditKeys}
            onGenerateValue={generate16BytesKey}
            tooltipId={tooltipIds.APP_SESSION_KEY}
          />
          <Form.Field
            mayGenerateValue
            title={lwVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
            name="session.keys.f_nwk_s_int_key.key"
            type="byte"
            min={16}
            max={16}
            required
            component={Input.Generate}
            onGenerateValue={generate16BytesKey}
            tooltipId={lwVersion >= 110 ? undefined : tooltipIds.NETWORK_SESSION_KEY}
          />
          {lwVersion >= 110 && (
            <Form.Field
              mayGenerateValue
              title={sharedMessages.sNwkSIKey}
              name="session.keys.s_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              required
              description={sharedMessages.sNwkSIKeyDescription}
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
            />
          )}
          {lwVersion >= 110 && (
            <Form.Field
              mayGenerateValue
              title={sharedMessages.nwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              required
              description={sharedMessages.nwkSEncKeyDescription}
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
            />
          )}
        </>
      )}
      <Form.Field
        required
        title={sharedMessages.devID}
        name="ids.device_id"
        placeholder={sharedMessages.deviceIdPlaceholder}
        component={Input}
        onFocus={handleIdTextSelect}
        inputRef={idInputRef}
        tooltipId={tooltipIds.DEVICE_ID}
        description={messages.deviceIdDescription}
      />
    </div>
  )
}

export { DeviceRegistrationFormSection as default, initialValues }
