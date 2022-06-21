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

import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

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

const DeviceTypeFormSection = props => {
  const [isClaiming, setIsClaiming] = useState(undefined)

  const onIdPrefill = useCallback(() => {
    // Reminder that DevEUI is used as id default (on blur).
    console.log('blur')
  })

  return (
    <>
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
      {isClaiming === true && <DeviceClaimingFormSection />}
      {isClaiming === false && <DeviceRegistrationFormSection />}
      {isClaiming !== undefined && (
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={isClaiming === true ? m.claimEndDevice : m.registerEndDevice}
          />
        </SubmitBar>
      )}
    </>
  )
}

export { DeviceTypeFormSection as default, initialValues }
