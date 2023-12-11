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
import { useNavigate } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import Form, { useFormContext } from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import Radio from '@ttn-lw/components/radio-button'
import Notification from '@ttn-lw/components/notification'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectAsEnabled, selectJsEnabled, selectNsEnabled } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { checkFromState } from '@account/lib/feature-checks'
import { mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { createDevice } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import m from './messages'
import { mayProvisionDevice, REGISTRATION_TYPES } from './utils'
import DeviceProvisioningFormSection, {
  initialValues as provisioningInitialValues,
} from './provisioning-form-section'
import DeviceTypeFormSection, { initialValues as typeInitialValues } from './type-form-section'
import DeviceQRScanFormSection from './qr-scan-form-section'
import validationSchema from './validation-schema'

const deviceOnboardingEnabled = selectNsEnabled() && selectAsEnabled() && selectJsEnabled()

const initialValues = merge(
  {
    _registration: REGISTRATION_TYPES.SINGLE,
    _withQRdata: false,
  },
  provisioningInitialValues,
  typeInitialValues,
)

const DeviceOnboardingFormInner = () => {
  const { values } = useFormContext()
  const template = useSelector(selectDeviceTemplate)

  // Submitting is allowed once the device type was specified and the claimability was determined.
  const maySubmit = values._claim !== null && mayProvisionDevice(values, template)

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
            <Form.Submit message={sharedMessages.registerEndDevice} component={SubmitButton} />
          </SubmitBar>
        </>
      )}
    </>
  )
}

const DeviceOnboardingForm = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const mayEditKeys = useSelector(state => checkFromState(mayEditApplicationDeviceKeys, state))
  const appId = useSelector(state => selectSelectedApplicationId(state))
  const validationContext = React.useMemo(() => ({ mayEditKeys }), [mayEditKeys])
  const [error, setError] = useState(undefined)

  const navigateToDevice = useCallback(
    (appId, deviceId) => navigate(`/applications/${appId}/devices/${deviceId}`),
    [navigate],
  )

  const handleSubmit = useCallback(
    async (values, { resetForm }, cleanedValues) => {
      const { _registration, _claim } = values
      const applicationIds = { application_id: appId }
      const { mac_state = {} } = cleanedValues
      const isClaiming = _claim === true

      setError()

      // Set the application ID.
      cleanedValues.ids.application_ids = applicationIds

      try {
        //  Do not attempt to set empty `current_parameters` on device creation.
        if (
          mac_state.current_parameters &&
          Object.keys(mac_state.current_parameters).length === 0
        ) {
          delete mac_state.current_parameters
        }

        // Obtain the device ID from authenticated identifiers when claiming.
        if (isClaiming) {
          cleanedValues.ids = {
            ...cleanedValues.ids,
            dev_eui: values.authenticated_identifiers.dev_eui,
            device_id: values.target_device_id,
          }
        }

        const endDevice = await dispatch(attachPromise(createDevice(appId, cleanedValues)))
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
                _registration: REGISTRATION_TYPES.MULTIPLE,
                // Reset registration values.
                ids: {
                  ...initialValues.ids,
                  join_eui: values.ids.join_eui,
                },
                frequency_plan_id: values.frequency_plan_id,
                version_ids: values.version_ids,
                // Reset claim values.
                authenticated_identifiers: {
                  ...initialValues.authenticated_identifiers,
                  join_eui: values.authenticated_identifiers.join_eui,
                },
                target_device_id: initialValues.target_device_id,
                // Only reset session when not using multicast.
                session: values.multicast ? values.session : initialValues.session,
              },
            })
            break
          case REGISTRATION_TYPES.SINGLE:
          default:
            navigateToDevice(appId, endDevice.ids.device_id)
        }
      } catch (error) {
        setError(error)
      }
    },
    [appId, dispatch, navigateToDevice],
  )

  return (
    <Form
      onSubmit={handleSubmit}
      error={error}
      initialValues={initialValues}
      validationSchema={validationSchema}
      validationContext={validationContext}
      validateSync={false}
      disabled={!deviceOnboardingEnabled}
    >
      {deviceOnboardingEnabled && <DeviceQRScanFormSection />}
      {!deviceOnboardingEnabled && (
        <Notification content={m.onboardingDisabled} className="mt-ls-m" warning small />
      )}
      <DeviceOnboardingFormInner />
    </Form>
  )
}

export default DeviceOnboardingForm
