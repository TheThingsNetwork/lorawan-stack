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

import React, { useCallback } from 'react'
import { merge } from 'lodash'
import { useSelector } from 'react-redux'

import Form, { useFormContext } from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { checkFromState } from '@account/lib/feature-checks'
import { mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import DeviceProvisioningFormSection, {
  initialValues as provisioningInitialValues,
} from './device-provisioning-form-section'
import DeviceTypeFormSection, {
  initialValues as typeInitialValues,
} from './device-type-form-section'
import validationSchema from './validation-schema'

const initialValues = merge({}, provisioningInitialValues, typeInitialValues)

const DeviceOnboardingFormInner = () => {
  const {
    values: {
      frequency_plan_id,
      lorawan_version,
      lorawan_phy_version,
      ids: { join_eui },
    },
  } = useFormContext()
  const maySubmit =
    Boolean(frequency_plan_id) &&
    Boolean(lorawan_version) &&
    Boolean(lorawan_phy_version) &&
    join_eui.length === 16

  return (
    <>
      <DeviceTypeFormSection />
      <DeviceProvisioningFormSection />
      {maySubmit && (
        <SubmitBar>
          <Form.Submit message={sharedMessages.addDevice} component={SubmitButton} />
        </SubmitBar>
      )}
    </>
  )
}

const DeviceOnboardingForm = () => {
  const mayEditKeys = useSelector(state => checkFromState(mayEditApplicationDeviceKeys, state))
  const validationContext = React.useMemo(() => ({ mayEditKeys }), [mayEditKeys])

  const handleSubmit = useCallback((values, formikBag, cleanedValues) => {
    console.log(values, cleanedValues)
    return Promise.resolve()
  }, [])

  return (
    <Form
      onSubmit={handleSubmit}
      initialValues={initialValues}
      hiddenFields={['network_server_address', 'application_server_address', 'join_server_address']}
      validationSchema={validationSchema}
      validationContext={validationContext}
      validateAgainstCleanedValues
    >
      <DeviceOnboardingFormInner />
    </Form>
  )
}

export default DeviceOnboardingForm
