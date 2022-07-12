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
import { connect, useSelector } from 'react-redux'
import { merge } from 'lodash'
import { push } from 'connected-react-router'

import tts from '@console/api/tts'

import Form from '@ttn-lw/components/form'
import toast from '@ttn-lw/components/toast'

import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectJsConfig, selectNsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'

import { checkFromState, mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { getApplicationDevEUICount, issueDevEUI } from '@console/store/actions/applications'
import { getTemplate } from '@console/store/actions/device-repository'

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

const initialValues = merge({}, provisioningInitialValues, typeInitialValues)

const DeviceOnboardingForm = props => {
  const {
    appId,
    getRegistrationTemplate,
    fetchDevEUICounter,
    issueDevEUI,
    navigateToDevice,
    mayEditKeys,
    createDevice,
  } = props
  const [isClaiming, setIsClaiming] = useState(undefined)

  const jsConfig = useSelector(selectJsConfig)
  const nsConfig = useSelector(selectNsConfig)
  const asConfig = useSelector(selectAsConfig)
  const asEnabled = asConfig.enabled
  const jsEnabled = jsConfig.enabled
  const nsEnabled = nsConfig.enabled
  const asUrl = asEnabled ? asConfig.base_url : undefined
  const jsUrl = jsEnabled ? jsConfig.base_url : undefined
  const nsUrl = nsEnabled ? nsConfig.base_url : undefined

  const validationContext = React.useMemo(
    () => ({
      mayEditKeys,
      asUrl,
      asEnabled,
      jsUrl,
      jsEnabled,
      nsUrl,
      nsEnabled,
    }),
    [asEnabled, asUrl, jsEnabled, jsUrl, mayEditKeys, nsEnabled, nsUrl],
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
        const device = await tts.DeviceClaim.claim(appId, qr_code, authenticatedIdentifiers)
        const { ids } = device
        await navigateToDevice(appId, ids.device_id)
      } catch (error) {
        setSubmitting(false)
      }
    },
    [appId, navigateToDevice],
  )

  const handleRegister = useCallback(
    async (values, setSubmitting, resetForm, setErrors) => {
      try {
        const { _registration, _inputMethod, ...castedValues } = validationSchema.cast(values, {
          context: validationContext,
        })

        const { ids, supports_join, mac_state = {} } = castedValues
        ids.application_ids = { application_id: appId }

        //  Do not attempt to set empty current_parameters on device creation.
        if (
          mac_state.current_parameters &&
          Object.keys(mac_state.current_parameters).length === 0
        ) {
          delete mac_state.current_parameters
        }

        await createDevice(appId, castedValues)
        switch (_registration) {
          case REGISTRATION_TYPES.MULTIPLE:
            toast({
              type: toast.types.SUCCESS,
              message: m.createSuccess,
            })
            resetForm({
              errors: {},
              values: {
                ...castedValues,
                ...initialValues,
                ids: {
                  ...initialValues.ids,
                  join_eui: supports_join ? castedValues.ids.join_eui : undefined,
                },
                frequency_plan_id: castedValues.frequency_plan_id,
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
    [appId, createDevice, validationContext, navigateToDevice],
  )

  const handleSubmit = useCallback(
    async (values, { setSubmitting, resetForm, setErrors }) => {
      if (isClaiming) {
        return handleClaim(values, setSubmitting, resetForm, setErrors)
      }

      return handleRegister(values, setSubmitting, resetForm, setErrors)
    },
    [handleClaim, handleRegister, isClaiming],
  )

  return (
    <Form
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
      stripUnusedFields
    >
      <DeviceTypeFormSection appId={appId} getRegistrationTemplate={getRegistrationTemplate} />
      <DeviceProvisioningFormSection
        appId={appId}
        fetchDevEUICounter={fetchDevEUICounter}
        issueDevEUI={issueDevEUI}
        mayEditKeys={mayEditKeys}
        isClaiming={isClaiming}
        setIsClaiming={setIsClaiming}
      />
    </Form>
  )
}

DeviceOnboardingForm.propTypes = {
  appId: PropTypes.string,
  createDevice: PropTypes.func.isRequired,
  fetchDevEUICounter: PropTypes.func,
  getRegistrationTemplate: PropTypes.func,
  issueDevEUI: PropTypes.func,
  mayEditKeys: PropTypes.bool.isRequired,
  navigateToDevice: PropTypes.func.isRequired,
}

DeviceOnboardingForm.defaultProps = {
  appId: undefined,
  getRegistrationTemplate: () => null,
  fetchDevEUICounter: () => null,
  issueDevEUI: () => null,
}

export default connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    mayEditKeys: checkFromState(mayEditApplicationDeviceKeys, state),
    createDevice: (appId, device) => tts.Applications.Devices.create(appId, device),
  }),
  dispatch => ({
    navigateToDevice: (appId, deviceId) =>
      dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
    fetchDevEUICounter: appId => dispatch(getApplicationDevEUICount(appId)),
    issueDevEUI: appId => dispatch(attachPromise(issueDevEUI(appId))),
    getRegistrationTemplate: (appId, version) => dispatch(getTemplate(appId, version)),
  }),
)(DeviceOnboardingForm)
