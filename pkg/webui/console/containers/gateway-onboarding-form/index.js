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

import GatewayApiKeysModal from '@console/components/gateway-api-keys-modal'

import { composeDataUri, downloadDataUriAsFile } from '@ttn-lw/lib/data-uri'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { createGateway, updateGateway } from '@console/store/actions/gateways'
import { createGatewayApiKey } from '@console/store/actions/api-keys'

import { selectUserId } from '@console/store/selectors/user'

import GatewayProvisioningFormSection from './gateway-provisioning-form'
import validationSchema from './gateway-provisioning-form/validation-schema'
import { initialValues as registerInitialValues } from './gateway-provisioning-form/gateway-registration-form-section'

const GatewayOnboardingForm = props => {
  const { onSuccess } = props
  const userId = useSelector(selectUserId)
  const dispatch = useDispatch()
  const [error, setError] = useState()
  const [cupsKey, setCupsKey] = useState()
  const [lnsKey, setLnsKey] = useState()
  const [modalVisible, setModalVisible] = useState(false)
  const [gatewayId, setGatewayId] = useState('')

  const initialValues = useMemo(
    () =>
      merge(
        {
          _ownerId: userId,
          enforce_duty_cycle: true,
          schedule_anytime_delay: '0.530s',
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
      const { _ownerId, _create_api_key_cups, _create_api_key_lns } = values

      const isUserOwner = _ownerId ? userId === _ownerId : true
      const ownerId = _ownerId ? _ownerId : userId
      const gatewayId = cleanValues.ids.gateway_id
      setGatewayId(gatewayId)

      try {
        await dispatch(attachPromise(createGateway(ownerId, cleanValues, isUserOwner)))
        if (_create_api_key_cups) {
          const generatedCupsKey = await generateCUPSApiKey(gatewayId)
          setCupsKey(generatedCupsKey.key)
        }
        if (_create_api_key_lns) {
          const generatedLnsKey = await generateLNSApiKey(gatewayId)
          const lbs_lns_secret = { value: btoa(generatedLnsKey.key) }
          await dispatch(attachPromise(updateGateway(gatewayId, { lbs_lns_secret })))
          setLnsKey(generatedLnsKey.key)
        }
        if (_create_api_key_cups || _create_api_key_lns) {
          return setModalVisible(true)
        }

        return onSuccess(gatewayId)
      } catch (error) {
        setError(error)
      }
    },
    [dispatch, generateCUPSApiKey, generateLNSApiKey, userId, onSuccess],
  )

  const handleSubmit = useCallback(
    (values, _, cleanValues) => handleRegistrationSubmit(values, cleanValues),
    [handleRegistrationSubmit],
  )

  const downloadLns = useCallback(() => {
    if (lnsKey) {
      const content = composeDataUri(`Authorization: Bearer ${lnsKey}\r\n`)
      downloadDataUriAsFile(content, 'tc.key')
    }
  }, [lnsKey])

  const downloadCups = useCallback(() => {
    if (cupsKey) {
      const content = composeDataUri(`Authorization: Bearer ${cupsKey}\r\n`)
      downloadDataUriAsFile(content, 'cups.key')
    }
  }, [cupsKey])

  const modalApprove = useCallback(() => onSuccess(gatewayId), [onSuccess, gatewayId])

  return (
    <>
      <GatewayApiKeysModal
        modalVisible={modalVisible}
        lnsKey={lnsKey}
        cupsKey={cupsKey}
        downloadLns={downloadLns}
        downloadCups={downloadCups}
        modalApprove={modalApprove}
      />
      <Form
        error={error}
        onSubmit={handleSubmit}
        initialValues={initialValues}
        hiddenFields={['gateway_server_address', 'enforce_duty_cycle', 'schedule_anytime_delay']}
        validationSchema={validationSchema}
        validateAgainstCleanedValues
      >
        <GatewayProvisioningFormSection userId={userId} />
      </Form>
    </>
  )
}

GatewayOnboardingForm.propTypes = {
  onSuccess: PropTypes.func.isRequired,
}

export default GatewayOnboardingForm
