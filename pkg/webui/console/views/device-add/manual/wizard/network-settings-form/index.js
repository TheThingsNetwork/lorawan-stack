// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import Wizard, { useWizardContext } from '@ttn-lw/components/wizard'
import Form from '@ttn-lw/components/form'

import PhyVersionInput from '@console/components/phy-version-input'
import MacSettingsSection from '@console/components/mac-settings-section'

import DevAddrInput from '@console/containers/dev-addr-input'
import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  ACTIVATION_MODES,
  LORAWAN_VERSIONS,
  parseLorawanMacVersion,
  generate16BytesKey,
} from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const excludePaths = ['_device_classes', 'class_b', 'class_c']

const defaultFormValues = {
  lorawan_phy_version: '',
  frequency_plan_id: '',
  mac_settings: {
    resets_f_cnt: false,
    supports_32_bit_f_cnt: true,
    ping_slot_periodicity: '',
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
  const { activationMode, lorawanVersion } = props
  const { error, snapshot } = useWizardContext()

  const [isClassB, setClassB] = React.useState(snapshot.supports_class_b)
  const handleClassBChange = React.useCallback(evt => {
    const { checked } = evt.target

    setClassB(checked)
  }, [])

  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const lwVersion = parseLorawanMacVersion(lorawanVersion)
  // Expand the advanced settings section:
  // 1. For multicast end devices becaise of the required `mac_settings.ping_slot_periodicity` field.
  // 2. For failed NS submission because of any possibly required`mac_settings` field.
  const expandAdvancedSettings = isMulticast || Boolean(error)

  const validationContext = React.useMemo(
    () => ({
      isClassB,
      activationMode,
    }),
    [activationMode, isClassB],
  )

  const initialFormValues = React.useMemo(
    () => validationSchema.cast(defaultFormValues, { context: validationContext }),
    [validationContext],
  )

  return (
    <Wizard.Form
      initialValues={initialFormValues}
      validationSchema={validationSchema}
      validationContext={validationContext}
      excludePaths={excludePaths}
    >
      <NsFrequencyPlansSelect
        required
        autoFocus
        glossaryId={glossaryIds.FREQUENCY_PLAN}
        name="frequency_plan_id"
      />
      <Form.Field
        required
        disabled
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={Select}
        options={LORAWAN_VERSIONS}
        glossaryId={glossaryIds.LORAWAN_VERSION}
      />
      <Form.Field
        required
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={PhyVersionInput}
        lorawanVersion={lorawanVersion}
        glossaryId={glossaryIds.REGIONAL_PARAMETERS}
      />
      <Form.Field
        title={sharedMessages.lorawanClassCapabilities}
        name="_device_classes"
        component={Checkbox.Group}
        required={isMulticast}
        glossaryId={glossaryIds.CLASSES}
      >
        <Checkbox
          name="class_b"
          label={sharedMessages.supportsClassB}
          onChange={handleClassBChange}
        />
        <Checkbox name="class_c" label={sharedMessages.supportsClassC} />
      </Form.Field>
      {(isMulticast || isABP) && (
        <>
          <DevAddrInput title={sharedMessages.devAddr} name="session.dev_addr" required />
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
            glossaryId={
              lwVersion >= 110
                ? glossaryIds.NETWORK_SESSION_KEY
                : glossaryIds.FORWARDING_NETWORK_SESSION_INTEGRITY_KEY
            }
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
              glossaryId={glossaryIds.SERVING_NETWORK_SESSION_INTEGRITY_KEY}
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
              glossaryId={glossaryIds.NETWORK_SESSION_ENCRYPTION_KEY}
            />
          )}
        </>
      )}
      <MacSettingsSection
        activationMode={activationMode}
        isClassB={isClassB}
        initiallyCollapsed={!expandAdvancedSettings}
      />
    </Wizard.Form>
  )
}

NetworkSettingsForm.propTypes = {
  activationMode: PropTypes.string.isRequired,
  lorawanVersion: PropTypes.string.isRequired,
}

const WrappedNetworkSettingsForm = withBreadcrumb('device.add.steps.network', props => (
  <Breadcrumb path={props.match.url} content={props.title} />
))(NetworkSettingsForm)

WrappedNetworkSettingsForm.propTypes = {
  match: PropTypes.match.isRequired,
  title: PropTypes.message.isRequired,
}

export default WrappedNetworkSettingsForm
