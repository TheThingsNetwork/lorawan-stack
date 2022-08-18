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

import Form, { useFormContext } from '@ttn-lw/components/form'

import LorawanVersionInput from '@console/components/lorawan-version-input'
import PhyVersionInput from '@console/components/phy-version-input'

import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { parseLorawanMacVersion, LORAWAN_VERSION_PAIRS } from '@console/lib/device-utils'

import AdvancedSettingsSection, {
  initialValues as advancedSettingsInitialValues,
} from './advanced-settings-section'

const initialValues = {
  lorawan_version: '',
  lorawan_phy_version: '',
  frequency_plan_id: '',
  ...advancedSettingsInitialValues,
}

// Always reset LW and PHY version when changing FP do avoid invalid
// version combinations that can otherwise occur.
const frequencyPlanValueSetter = ({ setValues, setFieldTouched }, { value }) => {
  setFieldTouched('lorawan_version', false)
  setFieldTouched('lorawan_phy_version', false)
  return setValues(values => ({
    ...values,
    frequency_plan_id: value,
    lorawan_version: '',
    lorawan_phy_version: '',
  }))
}

// Always reset the PHY version when setting the lorawan version to avoid
// invalid version combinations that would otherwise briefly occur until
// the PHY version is set by the field itself.
const lorawanVersionValueSetter = ({ setValues, setFieldTouched }, { value }) => {
  const phyVersions = LORAWAN_VERSION_PAIRS[parseLorawanMacVersion(value)] || []
  setFieldTouched('lorawan_phy_version', false)
  return setValues(values => ({
    ...values,
    lorawan_version: value,
    lorawan_phy_version: phyVersions.length === 1 ? phyVersions[0].value : '',
  }))
}

const DeviceTypeManualFormSection = () => {
  const {
    values: { frequency_plan_id, lorawan_version },
  } = useFormContext()

  return (
    <>
      <NsFrequencyPlansSelect
        required
        tooltipId={tooltipIds.FREQUENCY_PLAN}
        name="frequency_plan_id"
        valueSetter={frequencyPlanValueSetter}
      />
      <Form.Field
        required
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={LorawanVersionInput}
        tooltipId={tooltipIds.LORAWAN_VERSION}
        frequencyPlan={frequency_plan_id}
        valueSetter={lorawanVersionValueSetter}
      />
      <Form.Field
        required
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={PhyVersionInput}
        tooltipId={tooltipIds.REGIONAL_PARAMETERS}
        lorawanVersion={lorawan_version}
      />
      <AdvancedSettingsSection />
    </>
  )
}

export { DeviceTypeManualFormSection as default, initialValues }
