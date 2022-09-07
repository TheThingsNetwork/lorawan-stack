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
import DeviceQRScanFormSection from './device-qr-scan-form-section'
import validationSchema from './validation-schema'

const initialValues = merge(
  {
    _registration: REGISTRATION_TYPES.SINGLE,
    _withQRdata: false,
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

  const handleSubmit = useCallback(
    async (values, { resetForm }, cleanedValues) => {
      const { _registration } = values
      const applicationIds = { application_id: appId }
      const { authenticated_identifiers, target_device_id, ...submitValues } = cleanedValues
      const { mac_state = {} } = submitValues

      setError()

      // Set the application ID.
      submitValues.ids.application_ids = applicationIds

      try {
        // Initiate claiming, if the device is claimable.
        if (values._claim) {
          const claimValues = {
            authenticated_identifiers,
            target_device_id,
          }
          // Provision (claim) the device on the Join Server.
          const deviceIds = await dispatch(
            attachPromise(claimDevice(appId, undefined, claimValues)),
          )
          // Apply the resulting IDs to the submit values.
          submitValues.ids = deviceIds
        }

        //  Do not attempt to set empty `current_parameters` on device creation.
        if (
          mac_state.current_parameters &&
          Object.keys(mac_state.current_parameters).length === 0
        ) {
          delete mac_state.current_parameters
        }

        // Create the device in the stack (skipping JS, if claimed previously).
        await dispatch(attachPromise(createDevice(appId, submitValues)))
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
            navigateToDevice(appId, submitValues.ids.device_id)
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
    >
      <DeviceQRScanFormSection />
      <hr />
      <DeviceOnboardingFormInner />
    </Form>
  )
}

export default DeviceOnboardingForm
