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

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import Form from '@ttn-lw/components/form'
import Notification from '@ttn-lw/components/notification'
import Checkbox from '@ttn-lw/components/checkbox'

import PhyVersionInput from '@console/components/phy-version-input'
import MacSettingsSection from '@console/components/mac-settings-section'

import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'
import DevAddrInput from '@console/containers/dev-addr-input'

import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'
import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  parseLorawanMacVersion,
  ACTIVATION_MODES,
  LORAWAN_VERSIONS,
  generate16BytesKey,
} from '@console/lib/device-utils'

import messages from '../messages'
import {
  isDeviceABP,
  isDeviceMulticast,
  hasExternalJs,
  isDeviceJoined,
  isDeviceOTAA,
} from '../utils'

import validationSchema from './validation-schema'

const NetworkServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess, mayEditKeys, mayReadKeys } = props
  const { multicast = false, supports_join = false, supports_class_b = false } = device

  const isABP = isDeviceABP(device)
  const isMulticast = isDeviceMulticast(device)
  const isJoinedOTAA = isDeviceOTAA(device) && isDeviceJoined(device)

  const formRef = React.useRef(null)

  const [error, setError] = React.useState('')

  const [lorawanVersion, setLorawanVersion] = React.useState(device.lorawan_version)
  const lwVersion = parseLorawanMacVersion(lorawanVersion)

  const [isClassB, setClassB] = React.useState(supports_class_b)
  const handleClassBChange = React.useCallback(evt => {
    const { checked } = evt.target

    setClassB(checked)
  }, [])

  const initialActivationMode = supports_join
    ? ACTIVATION_MODES.OTAA
    : multicast
    ? ACTIVATION_MODES.MULTICAST
    : ACTIVATION_MODES.ABP

  const validationContext = React.useMemo(
    () => ({
      mayEditKeys,
      mayReadKeys,
      isJoined: isDeviceOTAA(device) && isDeviceJoined(device),
      externalJs: hasExternalJs(device),
    }),
    [device, mayEditKeys, mayReadKeys],
  )

  const initialValues = React.useMemo(
    () =>
      validationSchema.cast(
        {
          ...device,
          _activation_mode: initialActivationMode,
          _device_classes: { class_b: device.supports_class_b, class_c: device.supports_class_c },
        },
        { context: validationContext },
      ),
    [device, initialActivationMode, validationContext],
  )

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values, { context: validationContext })
      const updatedValues = diff(initialValues, castedValues, [
        '_activation_mode',
        'class_b',
        'class_c',
        'mac_settings',
      ])

      setError('')
      try {
        // Always submit current `mac_settings` values to avoid overwriting nested entries.
        await onSubmit({ ...updatedValues, mac_settings: castedValues.mac_settings })
        resetForm({ values: castedValues })
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit, onSubmitSuccess, validationContext],
  )

  const handleVersionChange = React.useCallback(
    version => {
      const isABP = initialValues._activation_mode === ACTIVATION_MODES.ABP
      const lwVersion = parseLorawanMacVersion(version)
      setLorawanVersion(version)
      const { setValues, values: formValues } = formRef.current
      const { session = {} } = formValues
      const { session: initialSession } = initialValues
      if (lwVersion >= 110) {
        const updatedSession = isABP
          ? {
              dev_addr: session.dev_addr,
              keys: {
                ...session.keys,
                s_nwk_s_int_key:
                  session.keys.s_nwk_s_int_key || initialSession.keys.s_nwk_s_int_key,
                nwk_s_enc_key: session.keys.nwk_s_enc_key || initialSession.keys.nwk_s_enc_key,
              },
            }
          : session
        setValues({
          ...formValues,
          lorawan_version: version,
          session: updatedSession,
        })
      } else {
        const updatedSession = isABP
          ? {
              dev_addr: session.dev_addr,
              keys: {
                f_nwk_s_int_key: session.keys.f_nwk_s_int_key,
              },
            }
          : session
        setValues({
          ...formValues,
          lorawan_version: version,
          session: updatedSession,
        })
      }
    },
    [initialValues],
  )

  // Notify the user that the session keys might be there, but since there are
  // no rights to read the keys we cannot display them.
  const showResetNotification = !mayReadKeys && mayEditKeys && !Boolean(device.session)

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
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={Select}
        required
        options={LORAWAN_VERSIONS}
        onChange={handleVersionChange}
        glossaryId={glossaryIds.LORAWAN_VERSION}
      />
      <Form.Field
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={PhyVersionInput}
        lorawanVersion={lorawanVersion}
        required
        glossaryId={glossaryIds.REGIONAL_PARAMETERS}
      />
      <NsFrequencyPlansSelect name="frequency_plan_id" required />
      <Form.Field
        title={sharedMessages.lorawanClassCapabilities}
        name="_device_classes"
        component={Checkbox.Group}
        required={isMulticast}
      >
        <Checkbox
          name="class_b"
          label={sharedMessages.supportsClassB}
          onChange={handleClassBChange}
        />
        <Checkbox name="class_c" label={sharedMessages.supportsClassC} />
      </Form.Field>
      <Form.Field
        title={sharedMessages.activationMode}
        disabled
        required
        name="_activation_mode"
        component={Radio.Group}
      >
        <Radio label={sharedMessages.otaa} value={ACTIVATION_MODES.OTAA} />
        <Radio label={sharedMessages.abp} value={ACTIVATION_MODES.ABP} />
        <Radio label={sharedMessages.multicast} value={ACTIVATION_MODES.MULTICAST} />
      </Form.Field>
      {(isABP || isMulticast || isJoinedOTAA) && (
        <>
          {showResetNotification && <Notification content={messages.keysResetWarning} info small />}
          <DevAddrInput
            title={sharedMessages.devAddr}
            name="session.dev_addr"
            disabled={!mayEditKeys}
            required={mayReadKeys && mayEditKeys}
          />
          <Form.Field
            title={lwVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
            name="session.keys.f_nwk_s_int_key.key"
            type="byte"
            min={16}
            max={16}
            disabled={!mayEditKeys}
            component={Input.Generate}
            mayGenerateValue={mayEditKeys}
            onGenerateValue={generate16BytesKey}
            glossaryId={
              lwVersion >= 110
                ? glossaryIds.NETWORK_SESSION_KEY
                : glossaryIds.FORWARDING_NETWORK_SESSION_INTEGRITY_KEY
            }
          />
          {lwVersion >= 110 && (
            <Form.Field
              title={sharedMessages.sNwkSIKey}
              name="session.keys.s_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.sNwkSIKeyDescription}
              disabled={!mayEditKeys}
              component={Input.Generate}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
            />
          )}
          {lwVersion >= 110 && (
            <Form.Field
              title={sharedMessages.nwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.nwkSEncKeyDescription}
              disabled={!mayEditKeys}
              component={Input.Generate}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
            />
          )}
        </>
      )}
      <MacSettingsSection activationMode={initialActivationMode} isClassB={isClassB} />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

NetworkServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  mayReadKeys: PropTypes.bool.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default NetworkServerForm
