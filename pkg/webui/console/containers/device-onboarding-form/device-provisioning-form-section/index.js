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

import React from 'react'
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
  const { values, setFieldValue, setValues } = useFormContext()
  const {
    _claim,
    _inputMethod,
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
  const mayConfirm = ids.join_eui.length === 16

  const resetJoinEui = React.useCallback(async () => {
    setValues(values => ({
      ...values,
      ids: {
        ...values.ids,
        join_eui: '',
      },
      _claim: '',
    }))
  }, [setValues])

  const getClaiming = React.useCallback(async () => {
    const claim = await dispatch(attachPromise(getInfoByJoinEUI({ join_eui: ids?.join_eui })))
    const supportsClaiming = claim.supports_claiming ?? false

    setFieldValue('_claim', supportsClaiming)
  }, [ids, setFieldValue, dispatch])

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
          component={Input}
          tooltipId={tooltipIds.JOIN_EUI}
          encode={joinEuiEncoder}
          decode={joinEuiDecoder}
        >
          {_claim === undefined ? (
            <Button disabled={!mayConfirm} onClick={getClaiming} message={msg.confirm} raw />
          ) : (
            <Button onClick={resetJoinEui} message={msg.reset} raw />
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
      {isClaiming && <DeviceClaimingFormSection />}
      {isRegistration && <DeviceRegistrationFormSection />}
    </>
  )
}

export { DeviceProvisioningFormSection as default, initialValues }
