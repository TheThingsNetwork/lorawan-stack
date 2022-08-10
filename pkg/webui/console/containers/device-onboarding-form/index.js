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
    values: { frequency_plan_id, lorawan_version, lorawan_phy_version, ids },
  } = useFormContext()

  const maySubmit =
    Boolean(frequency_plan_id) &&
    Boolean(lorawan_version) &&
    Boolean(lorawan_phy_version) &&
    ids?.join_eui?.length === 16

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

  const navigateToDevice = useCallback(
    (appId, deviceId) => dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
    [dispatch],
  )

  const handleClaim = useCallback(
    async (values, setSubmitting) => {
      const { ids, authentication_code, qr_code } = values

      let authenticatedIdentifiers
      if (!qr_code) {
        authenticatedIdentifiers = {
          join_eui: ids.join_eui,
          dev_eui: ids.dev_eui,
          authentication_code,
        }
      }

      try {
        const device = await dispatch(
          attachPromise(claimDevice(appId, qr_code, authenticatedIdentifiers)),
        )
        const { ids } = device
        navigateToDevice(appId, ids.device_id)
      } catch (error) {
        setSubmitting(false)
      }
    },
    [appId, navigateToDevice, dispatch],
  )

  const handleRegister = useCallback(
    async (values, setSubmitting, resetForm, setErrors, cleanedValues) => {
      try {
        const { _registration } = values
        const { ids, supports_join, mac_state = {} } = cleanedValues
        ids.application_ids = { application_id: appId }

        //  Do not attempt to set empty current_parameters on device creation.
        if (
          mac_state.current_parameters &&
          Object.keys(mac_state.current_parameters).length === 0
        ) {
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
      } catch (error) {
        setErrors(error)
        setSubmitting(false)
      }
    },
    [appId, dispatch, navigateToDevice],
  )

  const handleSubmit = useCallback(
    async (values, { setSubmitting, resetForm, setErrors }, cleanedValues) => {
      if (values._claim) {
        return handleClaim(cleanedValues, setSubmitting)
      }

      return handleRegister(values, setSubmitting, resetForm, setErrors, cleanedValues)
    },
    [handleClaim, handleRegister],
  )

  return (
    <Form
      onSubmit={handleSubmit}
      initialValues={initialValues}
      hiddenFields={[
        'network_server_address',
        'application_server_address',
        'join_server_address',
        'frequency_plan_id',
        'lorawan_phy_version',
        'lorawan_version',
        'supports_join',
      ]}
      validationSchema={validationSchema}
      validationContext={validationContext}
      validateAgainstCleanedValues
    >
      <DeviceOnboardingFormInner />
    </Form>
  )
}

export default DeviceOnboardingForm
