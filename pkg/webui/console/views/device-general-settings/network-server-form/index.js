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
import Checkbox from '../../../../components/checkbox'
import Input from '../../../../components/input'
import Radio from '../../../../components/radio-button'
import Select from '../../../../components/select'
import Form from '../../../../components/form'
import { NsFrequencyPlansSelect } from '../../../containers/freq-plans-select'
import DevAddrInput from '../../../containers/dev-addr-input'

import diff from '../../../../lib/diff'
import m from '../../../components/device-data-form/messages'
import sharedMessages from '../../../../lib/shared-messages'
import PropTypes from '../../../../lib/prop-types'

import { parseLorawanMacVersion, isDeviceABP, isDeviceMulticast, ACTIVATION_MODES } from '../utils'
import validationSchema from './validation-schema'

const lorawanVersions = [
  { value: '1.0.0', label: 'MAC V1.0' },
  { value: '1.0.1', label: 'MAC V1.0.1' },
  { value: '1.0.2', label: 'MAC V1.0.2' },
  { value: '1.0.3', label: 'MAC V1.0.3' },
  { value: '1.1.0', label: 'MAC V1.1' },
]

const lorawanPhyVersions = [
  { value: '1.0.0', label: 'PHY V1.0' },
  { value: '1.0.1', label: 'PHY V1.0.1' },
  { value: '1.0.2-a', label: 'PHY V1.0.2 REV A' },
  { value: '1.0.2-b', label: 'PHY V1.0.2 REV B' },
  { value: '1.0.3-a', label: 'PHY V1.0.3 REV A' },
  { value: '1.1.0-a', label: 'PHY V1.1 REV A' },
  { value: '1.1.0-b', label: 'PHY V1.1 REV B' },
]

const NetworkServerForm = React.memo(props => {
  const { device, onSubmit } = props

  const isABP = isDeviceABP(device)
  const isMulticast = isDeviceMulticast(device)

  const formRef = React.useRef(null)

  const [error, setError] = React.useState('')
  const [resetsFCnt, setResetsFCnt] = React.useState(
    (isABP && device.mac_settings && device.mac_settings.resets_f_cnt) || false,
  )
  const [lorawanVersion, setLorawanVersion] = React.useState(
    parseLorawanMacVersion(device.lorawan_version),
  )

  const initialValues = React.useMemo(() => {
    const {
      lorawan_version,
      lorawan_phy_version,
      frequency_plan_id,
      supports_class_c = false,
      supports_join = false,
      multicast = false,
      session,
      mac_settings = {},
    } = device

    let _activation_mode = ACTIVATION_MODES.ABP
    if (supports_join) {
      _activation_mode = ACTIVATION_MODES.OTAA
    } else if (multicast) {
      _activation_mode = ACTIVATION_MODES.MULTICAST
    }

    return {
      lorawan_version,
      lorawan_phy_version,
      frequency_plan_id,
      supports_class_c,
      session,
      _activation_mode,
      mac_settings,
    }
  }, [device])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const isABP = initialValues._activation_mode === ACTIVATION_MODES.ABP

      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, ['_activation_mode', '_external_js'])

      if (isABP) {
        // Do not reset session keys
        if (updatedValues.session.keys && Object.keys(updatedValues.session.keys).length === 0) {
          delete updatedValues.session.keys
        }

        if (Object.keys(updatedValues.session).length === 0) {
          delete updatedValues.session
        }
      }

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

  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  const handleVersionChange = React.useCallback(
    version => {
      const isABP = initialValues._activation_mode === ACTIVATION_MODES.ABP
      const lwVersion = parseLorawanMacVersion(version)
      setLorawanVersion(lwVersion)

      const { setValues, state: formState } = formRef.current
      const { session = {} } = formState.values
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
          ...formState.values,
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
          ...formState.values,
          lorawan_version: version,
          session: updatedSession,
        })
      }
    },
    [initialValues],
  )

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
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={Select}
        required
        options={lorawanVersions}
        onChange={handleVersionChange}
      />
      <Form.Field
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={Select}
        required
        options={lorawanPhyVersions}
      />
      <NsFrequencyPlansSelect name="frequency_plan_id" required />
      <Form.Field title={m.supportsClassC} name="supports_class_c" component={Checkbox} />
      <Form.Field
        title={m.activationMode}
        disabled
        required
        name="_activation_mode"
        component={Radio.Group}
        horizontal={false}
      >
        <Radio label={m.otaa} value={ACTIVATION_MODES.OTAA} />
        <Radio label={m.abp} value={ACTIVATION_MODES.ABP} />
        <Radio label={m.multicast} value={ACTIVATION_MODES.MULTICAST} />
      </Form.Field>
      {(isABP || isMulticast) && (
        <>
          <DevAddrInput
            title={sharedMessages.devAddr}
            name="session.dev_addr"
            placeholder={m.leaveBlankPlaceholder}
            description={m.deviceAddrDescription}
            required
          />
          <Form.Field
            title={lorawanVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
            name="session.keys.f_nwk_s_int_key.key"
            type="byte"
            min={16}
            max={16}
            placeholder={m.leaveBlankPlaceholder}
            description={lorawanVersion >= 110 ? m.fNwkSIntKeyDescription : m.nwkSKeyDescription}
            component={Input}
          />
          {lorawanVersion >= 110 && (
            <Form.Field
              title={sharedMessages.sNtwkSIKey}
              name="session.keys.s_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              placeholder={m.leaveBlankPlaceholder}
              description={m.sNtwkSIKeyDescription}
              component={Input}
            />
          )}
          {lorawanVersion >= 110 && (
            <Form.Field
              title={sharedMessages.ntwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              placeholder={m.leaveBlankPlaceholder}
              description={m.ntwkSEncKeyDescription}
              component={Input}
            />
          )}
          {!isMulticast && (
            <Form.Field
              title={m.resetsFCnt}
              onChange={handleResetsFCntChange}
              warning={resetsFCnt ? m.resetWarning : undefined}
              name="mac_settings.resets_f_cnt"
              component={Checkbox}
            />
          )}
        </>
      )}
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

NetworkServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  onSubmit: PropTypes.func.isRequired,
}

export default NetworkServerForm
