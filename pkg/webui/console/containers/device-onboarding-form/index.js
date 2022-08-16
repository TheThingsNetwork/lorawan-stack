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
import { push } from 'connected-react-router'
import { useDispatch, useSelector } from 'react-redux'

import Form, { useFormContext } from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import Radio from '@ttn-lw/components/radio-button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState } from '@account/lib/feature-checks'
import { mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { claimDevice } from '@console/store/actions/claim'
import { createDevice } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import m from './messages'
import { REGISTRATION_TYPES } from './utils'
import DeviceProvisioningFormSection, {
  initialValues as provisioningInitialValues,
} from './device-provisioning-form-section'
import DeviceTypeFormSection, {
  initialValues as typeInitialValues,
} from './device-type-form-section'
import validationSchema from './validation-schema'

const initialValues = merge(
  {
    _registration: REGISTRATION_TYPES.SINGLE,
  },
  provisioningInitialValues,
  typeInitialValues,
)

const DeviceOnboardingFormInner = () => {
  const {
    values: { frequency_plan_id, lorawan_version, lorawan_phy_version, _claim },
  } = useFormContext()

  const maySubmit =
    Boolean(frequency_plan_id) &&
    Boolean(lorawan_version) &&
    Boolean(lorawan_phy_version) &&
    typeof _claim === 'boolean'

  return (
    <>
      <DeviceTypeFormSection />
      <DeviceProvisioningFormSection />
      {maySubmit && (
        <>
          <Form.Field title={m.afterRegistration} name="_registration" component={Radio.Group}>
            <Radio label={m.singleRegistration} value={REGISTRATION_TYPES.SINGLE} />
            <Radio label={m.multipleRegistration} value={REGISTRATION_TYPES.MULTIPLE} />
          </Form.Field>
          <SubmitBar>
            <Form.Submit message={sharedMessages.addDevice} component={SubmitButton} />
          </SubmitBar>
        </>
      )}
    </>
  )
}

const DeviceOnboardingForm = () => {
  const dispatch = useDispatch()
  const mayEditKeys = useSelector(state => checkFromState(mayEditApplicationDeviceKeys, state))
  const appId = useSelector(state => selectSelectedApplicationId(state))
  const validationContext = React.useMemo(() => ({ mayEditKeys }), [mayEditKeys])
  const [error, setError] = useState(undefined)

  const navigateToDevice = useCallback(
    (appId, deviceId) => dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
    [dispatch],
  )

  const handleClaim = useCallback(
    async values => {
      const { qr_code, ids, ...rest } = values
      const device = await dispatch(attachPromise(claimDevice(appId, qr_code, rest)))
      const { deviceIds } = device
      navigateToDevice(appId, deviceIds.device_id)
    },
    [appId, navigateToDevice, dispatch],
  )

  const handleRegister = useCallback(
    async (values, resetForm, { authenticated_identifiers, ...cleanedValues }) => {
      const { _registration } = values
      const { ids, supports_join, mac_state = {} } = cleanedValues
      ids.application_ids = { application_id: appId }

      //  Do not attempt to set empty current_parameters on device creation.
      if (mac_state.current_parameters && Object.keys(mac_state.current_parameters).length === 0) {
        delete mac_state.current_parameters
      }

      await dispatch(attachPromise(createDevice(appId, cleanedValues)))
      switch (_registration) {
        case REGISTRATION_TYPES.MULTIPLE:
          toast({
            type: toast.types.SUCCESS,
            message: m.createSuccess,
          })
          resetForm({
            errors: {},
            values: {
              ...values,
              ...initialValues,
              ids: {
                ...initialValues.ids,
                join_eui: supports_join ? values.ids.join_eui : undefined,
              },
              frequency_plan_id: values.frequency_plan_id,
              version_ids: values.version_ids,
              _registration: REGISTRATION_TYPES.MULTIPLE,
            },
          })
          break
        case REGISTRATION_TYPES.SINGLE:
          resetForm({ values: initialValues })
          toast({
            type: toast.types.SUCCESS,
            message: m.createSuccess,
          })
        default:
          navigateToDevice(appId, ids.device_id)
      }
    },
    [appId, dispatch, navigateToDevice],
  )

  const handleSubmit = useCallback(
    async (values, { resetForm }, cleanedValues) => {
      try {
        let result
        if (values._claim) {
          result = await handleClaim(cleanedValues)
        } else {
          result = await handleRegister(values, resetForm, cleanedValues)
        }

        return result
      } catch (error) {
        setError(error)
      }
    },
    [handleClaim, handleRegister],
  )

  return (
    <Form
      onSubmit={handleSubmit}
      error={error}
      initialValues={initialValues}
      validationSchema={validationSchema}
      validationContext={validationContext}
      validateAgainstCleanedValues
    >
      <DeviceOnboardingFormInner />
    </Form>
  )
}

export default DeviceOnboardingForm
