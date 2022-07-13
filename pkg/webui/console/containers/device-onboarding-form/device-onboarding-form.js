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

import Form from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import DeviceProvisioningFormSection, {
  initialValues as provisioningInitialValues,
} from './device-provisioning-form-section'
import DeviceTypeFormSection, {
  initialValues as typeInitialValues,
} from './device-type-form-section'
import validationSchema from './validation-schema'

const initialValues = merge({}, provisioningInitialValues, typeInitialValues)

const DeviceOnboardingForm = props => {
  const { appId, getRegistrationTemplate } = props
  const handleSubmit = useCallback(async values => {
    console.log(values)
  }, [])

  return (
    <Form
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
      stripUnusedFields
    >
      <DeviceTypeFormSection appId={appId} getRegistrationTemplate={getRegistrationTemplate} />
      <DeviceProvisioningFormSection />
      <SubmitBar>
        <Form.Submit message={sharedMessages.addDevice} component={SubmitButton} />
      </SubmitBar>
    </Form>
  )
}

DeviceOnboardingForm.propTypes = {
  appId: PropTypes.string,
  getRegistrationTemplate: PropTypes.func,
}

DeviceOnboardingForm.defaultProps = {
  appId: undefined,
  getRegistrationTemplate: () => null,
}

export default DeviceOnboardingForm
