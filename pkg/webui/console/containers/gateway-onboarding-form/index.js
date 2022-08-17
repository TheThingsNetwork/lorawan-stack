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

import React, { useCallback, useMemo, useState } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { merge } from 'lodash'

import Form from '@ttn-lw/components/form'

import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { createGateway } from '@console/store/actions/gateways'
import { createGatewayApiKey } from '@console/store/actions/api-keys'

import { selectUserId } from '@console/store/selectors/logout'

import GatewayProvisioningFormSection from './gateway-provisioning-form'
import validationSchema from './gateway-provisioning-form/validation-schema'
import { initialValues as registerInitialValues } from './gateway-provisioning-form/gateway-registration-form-section'

const GatewayOnboardingForm = props => {
  const { onSuccess } = props
  const userId = useSelector(selectUserId)
  const dispatch = useDispatch()
  const [error, setError] = useState()

  const initialValues = useMemo(
    () =>
      merge(
        {
          _ownerId: userId,
          ids: { eui: '' },
        },
        registerInitialValues,
      ),
    [userId],
  )

  const generateCUPSApiKey = useCallback(
    gateway_id => {
      const key = {
        name: `cups-api-key-${Date.now()}`,
        rights: [
          'RIGHT_GATEWAY_INFO',
          'RIGHT_GATEWAY_SETTINGS_BASIC',
          'RIGHT_GATEWAY_READ_SECRETS',
        ],
      }

      return dispatch(attachPromise(createGatewayApiKey(gateway_id, key)))
    },
    [dispatch],
  )

  const generateLNSApiKey = useCallback(
    gateway_id => {
      const key = {
        name: `lns-api-key-${Date.now()}`,
        rights: ['RIGHT_GATEWAY_INFO', 'RIGHT_GATEWAY_LINK'],
      }

      return dispatch(attachPromise(createGatewayApiKey(gateway_id, key)))
    },
    [dispatch],
  )

  const handleRegistrationSubmit = useCallback(
    async (values, cleanValues) => {
      const { _owner_id, _create_api_key_cups, _create_api_key_lns } = values

      const isUserOwner = _owner_id ? userId === _owner_id : true
      const ownerId = _owner_id ? _owner_id : userId
      const gatewayId = cleanValues.ids.gateway_id

      try {
        await dispatch(attachPromise(createGateway(ownerId, cleanValues, isUserOwner)))
        if (_create_api_key_cups) {
          await generateCUPSApiKey(gatewayId)
        }
        if (_create_api_key_lns) {
          await generateLNSApiKey(gatewayId)
        }
        onSuccess(gatewayId)
      } catch (error) {
        setError(error)
      }
    },
    [dispatch, generateCUPSApiKey, generateLNSApiKey, onSuccess, userId],
  )

  const handleSubmit = React.useCallback(
    (values, _, cleanValues) => handleRegistrationSubmit(values, cleanValues),
    [handleRegistrationSubmit],
  )

  return (
    <Form
      error={error}
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
      validateAgainstCleanedValues
    >
      <GatewayProvisioningFormSection userId={userId} />
    </Form>
  )
}

GatewayOnboardingForm.propTypes = {
  onSuccess: PropTypes.func.isRequired,
}

export default GatewayOnboardingForm
