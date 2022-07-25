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

import React, { useCallback, useState } from 'react'
import { merge } from 'lodash'
import { useSelector } from 'react-redux'

import Input from '@ttn-lw/components/input'
import Form, { useFormContext } from '@ttn-lw/components/form'

import Message from '@ttn-lw/lib/components/message'

import messages from '@account/containers/profile-settings-form/messages'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import { hasValidDeviceRepositoryType } from '../utils'
import m from '../messages'

import DeviceClaimingFormSection, {
  initialValues as claimingInitialValues,
} from './device-claiming-form-section'
import DeviceRegistrationFormSection, {
  initialValues as registrationInitialValues,
} from './device-registration-form-section'

const initialValues = merge(
  {
    ids: {
      join_eui: '',
    },
  },
  claimingInitialValues,
  registrationInitialValues,
)

const DeviceTypeFormSection = () => {
  const [isClaiming, setIsClaiming] = useState(undefined)
  const {
    values: {
      _inputMethod,
      version_ids: version,
      frequency_plan_id,
      lorawan_version,
      lorawan_phy_version,
    },
  } = useFormContext()
  const template = useSelector(selectDeviceTemplate)
  const mayProvisionDevice =
    (Boolean(frequency_plan_id) && Boolean(lorawan_version) && Boolean(lorawan_phy_version)) ||
    (_inputMethod === 'device-repository' && hasValidDeviceRepositoryType(version, template))

  const onIdPrefill = useCallback(() => {
    // Reminder that DevEUI is used as id default (on blur).
    console.log('blur')
  })

  return (
    <>
      {mayProvisionDevice ? (
        <Form.Field
          title={sharedMessages.joinEUI}
          name="ids.join_eui"
          type="byte"
          min={8}
          max={8}
          required
          component={Input}
          tooltipId={tooltipIds.JOIN_EUI}
          onBlur={onIdPrefill}
        />
      ) : (
        <Message
          content={_inputMethod === 'device-repository' ? m.continueDeviceRepo : m.continueManual}
          className="mt-ls-m mb-ls-m"
          component="div"
        />
      )}
      {isClaiming === true && <DeviceClaimingFormSection />}
      {isClaiming === false && <DeviceRegistrationFormSection />}
    </>
  )
}

export { DeviceTypeFormSection as default, initialValues }
