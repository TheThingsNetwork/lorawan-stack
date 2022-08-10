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
    ids: {
      join_eui: '',
    },
    _claim: '',
  },
  claimingInitialValues,
  registrationInitialValues,
)

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
  const mayProvisionDevice =
    (Boolean(frequency_plan_id) && Boolean(lorawan_version) && Boolean(lorawan_phy_version)) ||
    (_inputMethod === 'device-repository' && hasValidDeviceRepositoryType(version, template))
  const mayConfirm = _claim === '' || _claim === undefined

  const isClaimable = React.useCallback(async () => {
    if (_claim !== '') {
      setValues(values => ({
        ...values,
        ids: {
          ...values.ids,
          join_eui: '',
        },
        _claim: '',
      }))
    } else {
      const claim = await dispatch(attachPromise(getInfoByJoinEUI({ join_eui: ids?.join_eui })))
      const supportsClaiming = claim.supports_claiming ?? false

      setFieldValue('_claim', supportsClaiming)
    }
  }, [ids, setFieldValue, _claim, dispatch, setValues])

  React.useEffect(() => {
    // Merge the new device template with other form values.
    if (template && hasValidDeviceRepositoryType && _inputMethod === 'device-repository') {
      const { end_device } = template
      setValues(merge(values, end_device), false)
    }
  }, [template, _inputMethod, setValues, values])

  return (
    <>
      {mayProvisionDevice ? (
        <Form.Field
          title={sharedMessages.joinEUI}
          name="ids.join_eui"
          type="byte"
          min={8}
          max={8}
          required
          component={Input}
          actionDisable={!Boolean(ids?.join_eui)}
          action={{
            type: 'button',
            disable: !Boolean(ids?.join_eui),
            title: mayConfirm ? msg.confirm : msg.reset,
            message: mayConfirm ? msg.confirm : msg.reset,
            onClick: isClaimable,
            raw: true,
          }}
          tooltipId={tooltipIds.JOIN_EUI}
        />
      ) : (
        <Message
          content={_inputMethod === 'device-repository' ? m.continueDeviceRepo : m.continueManual}
          className="mt-ls-m mb-ls-m"
          component="div"
        />
      )}
      {mayProvisionDevice && _claim === '' && (
        <Message content={msg.continue} className="mt-ls-m mb-ls-m" component="div" />
      )}
      {mayProvisionDevice && _claim === true && (
        <>
          <Message content={msg.claiming} className="mb-ls-s" component="div" />
          <DeviceClaimingFormSection />
        </>
      )}
      {mayProvisionDevice && _claim === false && (
        <>
          <Message content={msg.registration} className="mb-ls-s" component="div" />
          <DeviceRegistrationFormSection />
        </>
      )}
    </>
  )
}

export { DeviceProvisioningFormSection as default, initialValues }
