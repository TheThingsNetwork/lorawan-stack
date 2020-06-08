// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import Select from '@ttn-lw/components/select'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Wizard from '@ttn-lw/components/wizard'
import Form from '@ttn-lw/components/form'

import DevAddrInput from '@console/containers/dev-addr-input'
import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  ACTIVATION_MODES,
  LORAWAN_VERSIONS,
  LORAWAN_PHY_VERSIONS,
  parseLorawanMacVersion,
  generate16BytesKey,
} from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const defaultFormValues = {
  lorawan_phy_version: '',
  frequency_plan_id: '',
  supports_class_c: false,
  mac_settings: {
    resets_f_cnt: false,
  },
  session: {
    dev_addr: '',
    keys: {
      f_nwk_s_int_key: { key: '' },
      s_nwk_s_int_key: { key: '' },
      nwk_s_enc_key: { key: '' },
    },
  },
  supports_join: false,
  multicast: false,
}

const NetworkSettingsForm = props => {
  const { activationMode, lorawanVersion, error } = props

  const validationContext = React.useMemo(
    () => ({
      activationMode,
    }),
    [activationMode],
  )

  const [resetsFCnt, setResetsFCnt] = React.useState(false)
  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const lwVersion = parseLorawanMacVersion(lorawanVersion)

  return (
    <Wizard.Form
      initialValues={defaultFormValues}
      validationSchema={validationSchema}
      validationContext={validationContext}
      error={error}
    >
      <NsFrequencyPlansSelect required autoFocus name="frequency_plan_id" />
      <Form.Field
        required
        disabled
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={Select}
        options={LORAWAN_VERSIONS}
      />
      <Form.Field
        required
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={Select}
        options={LORAWAN_PHY_VERSIONS}
      />
      <Form.Field
        title={sharedMessages.supportsClassC}
        name="supports_class_c"
        component={Checkbox}
      />
      {(isMulticast || isABP) && (
        <>
          <DevAddrInput
            title={sharedMessages.devAddr}
            name="session.dev_addr"
            description={sharedMessages.deviceAddrDescription}
            required
          />
          {isABP && (
            <Form.Field
              title={sharedMessages.resetsFCnt}
              onChange={handleResetsFCntChange}
              warning={resetsFCnt ? sharedMessages.resetWarning : undefined}
              name="mac_settings.resets_f_cnt"
              component={Checkbox}
            />
          )}
          <Form.Field
            mayGenerateValue
            title={lwVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
            name="session.keys.f_nwk_s_int_key.key"
            type="byte"
            min={16}
            max={16}
            required
            description={
              lorawanVersion >= 110
                ? sharedMessages.fNwkSIntKeyDescription
                : sharedMessages.nwkSKeyDescription
            }
            component={Input.Generate}
            onGenerateValue={generate16BytesKey}
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
    </Wizard.Form>
  )
}

NetworkSettingsForm.propTypes = {
  activationMode: PropTypes.string.isRequired,
  error: PropTypes.error,
  lorawanVersion: PropTypes.string.isRequired,
}

NetworkSettingsForm.defaultProps = {
  error: undefined,
}

const WrappedNetworkSettingsForm = withBreadcrumb('device.add.steps.network', props => (
  <Breadcrumb path={props.match.url} content={props.title} />
))(NetworkSettingsForm)

WrappedNetworkSettingsForm.propTypes = {
  match: PropTypes.match.isRequired,
  title: PropTypes.message.isRequired,
}

export default WrappedNetworkSettingsForm
