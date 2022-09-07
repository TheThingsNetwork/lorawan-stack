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

import React, { useEffect, useCallback } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { defineMessages } from 'react-intl'
import { merge } from 'lodash'

import Input from '@ttn-lw/components/input'
import Form, { useFormContext } from '@ttn-lw/components/form'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getInfoByJoinEUI } from '@console/store/actions/claim'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import { hasValidDeviceRepositoryType } from '../utils'
import m from '../messages'

import DeviceClaimingFormSection, {
  initialValues as claimingInitialValues,
} from './device-claiming-form-section'
import DeviceRegistrationFormSection, {
  initialValues as registrationInitialValues,
} from './device-registration-form-section'

const msg = defineMessages({
  continue:
    'To continue, please enter the JoinEUI of the end device so we can determine onboarding options',
  confirm: 'Confirm',
  reset: 'Reset',
  registration: '✔ This end device can be registered on the network',
  claiming: '✔ This end device can be claimed from its current owner',
})

const initialValues = merge(
  {
    authenticated_identifiers: {
      join_eui: '',
    },
    ids: {
      join_eui: '',
    },
    _claim: undefined,
  },
  claimingInitialValues,
  registrationInitialValues,
)

// Save EUI in both fields.
const joinEuiEncoder = value => ({
  ids: { join_eui: value },
  authenticated_identifiers: {
    join_eui: value,
  },
})
const joinEuiDecoder = value => value?.ids?.join_eui || ''

const DeviceProvisioningFormSection = () => {
  const dispatch = useDispatch()
  const { values, setValues } = useFormContext()
  const {
    _claim,
    _inputMethod,
    _withQRdata,
    version_ids: version,
    frequency_plan_id,
    lorawan_version,
    lorawan_phy_version,
    ids,
  } = values

  const template = useSelector(selectDeviceTemplate)
  const isClaiming = _claim === true
  const isRegistration = _claim === false
  const isClaimingDetermined = _claim !== undefined
  const isRepository = _inputMethod === 'device-repository'
  let joinEuiConfirmationMessage

  if (isClaimingDetermined) {
    joinEuiConfirmationMessage = isClaiming ? msg.claiming : msg.registration
  }

  const mayProvisionDevice =
    (Boolean(frequency_plan_id) && Boolean(lorawan_version) && Boolean(lorawan_phy_version)) ||
    (_inputMethod === 'device-repository' && hasValidDeviceRepositoryType(version, template))
  const mayConfirm = ids?.join_eui?.length === 16

  const resetJoinEui = React.useCallback(() => {
    setValues(values => ({
      ...values,
      ids: {
        ...values.ids,
        join_eui: '',
      },
      _claim: undefined,
    }))
  }, [setValues])

  const handleJoinEuiConfirm = useCallback(async () => {
    const claim = await dispatch(attachPromise(getInfoByJoinEUI({ join_eui: ids?.join_eui })))
    const supportsClaiming = claim.supports_claiming ?? false

    setValues(values => ({
      ...values,
      _claim: supportsClaiming,
      // In case of claiming, the creation on the join server needs to be skipped.
      join_server_address: supportsClaiming ? undefined : values.join_server_address,
    }))
  }, [dispatch, ids.join_eui, setValues])

  const handleJoinEuiKeyDown = useCallback(
    event => {
      if (event.key === 'Enter' && mayConfirm) {
        event.preventDefault()
        event.target.blur()
        handleJoinEuiConfirm()
      }
    },
    [handleJoinEuiConfirm, mayConfirm],
  )

  useEffect(() => {
    // Auto-confirm the join EUI when using QR code data.
    if (_withQRdata) {
      handleJoinEuiConfirm()
    }
  }, [_withQRdata, handleJoinEuiConfirm, ids.join_eui.length])

  return (
    <>
      {mayProvisionDevice ? (
        <Form.Field
          title={sharedMessages.joinEUI}
          name="ids.join_eui,authenticated_identifiers.join_eui"
          description={joinEuiConfirmationMessage}
          type="byte"
          min={8}
          max={8}
          required
          disabled={isClaimingDetermined || _withQRdata}
          component={Input}
          tooltipId={tooltipIds.JOIN_EUI}
          encode={joinEuiEncoder}
          decode={joinEuiDecoder}
          onKeyDown={handleJoinEuiKeyDown}
        >
          {_claim === undefined ? (
            <Button
              type="button"
              disabled={!mayConfirm}
              onClick={handleJoinEuiConfirm}
              message={msg.confirm}
              className="ml-cs-xs"
            />
          ) : (
            <Button
              onClick={resetJoinEui}
              type="button"
              message={msg.reset}
              className="ml-cs-xs"
              disabled={_withQRdata}
            />
          )}
        </Form.Field>
      ) : (
        <Message
          content={isRepository ? m.continueDeviceRepo : m.continueManual}
          className="mt-ls-m mb-ls-m"
          component="div"
        />
      )}
      {!isClaimingDetermined && mayProvisionDevice && (
        <Message content={msg.continue} className="mt-ls-m mb-ls-m" component="div" />
      )}
      {mayProvisionDevice && isClaiming && <DeviceClaimingFormSection />}
      {mayProvisionDevice && isRegistration && <DeviceRegistrationFormSection />}
    </>
  )
}

export { DeviceProvisioningFormSection as default, initialValues }
