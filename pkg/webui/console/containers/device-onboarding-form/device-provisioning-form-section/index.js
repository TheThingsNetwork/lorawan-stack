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

import React, { useState } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import { merge } from 'lodash'

import tts from '@console/api/tts'

import Input from '@ttn-lw/components/input'
import Form, { useFormContext } from '@ttn-lw/components/form'
import Radio from '@ttn-lw/components/radio-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import m from '../messages'
import { REGISTRATION_TYPES } from '../utils'

import DeviceClaimingFormSection, {
  initialValues as claimingInitialValues,
} from './device-claiming-form-section'
import DeviceRegistrationFormSection, {
  initialValues as registrationInitialValues,
} from './device-registration-form-section'

const msg = defineMessages({
  joinEUIDesc:
    'To continue, please enter the JoinEUI of the end device so we can determine onboarding options',
  confirm: 'Confirm',
  reset: 'Reset',
  registration: '✔ This end device can be registered on the network',
  claiming: '✔ This end device can be claimed from its current owner',
})

const initialValues = merge(
  {
    ids: {
      join_eui: undefined,
    },
    _registration: REGISTRATION_TYPES.SINGLE,
  },
  claimingInitialValues,
  registrationInitialValues,
)

const DeviceProvisioningFormSection = props => {
  const { appId, fetchDevEUICounter, issueDevEUI, mayEditKeys, setIsClaiming, isClaiming } = props
  const { values, setFieldValue, setValues } = useFormContext()
  const { _inputMethod, version_ids } = values
  const [joinEUIDescription, setJoinEUIDescription] = useState(msg.joinEUIDesc)
  const template = useSelector(selectDeviceTemplate)

  const isClaimable = React.useCallback(async () => {
    if (isClaiming !== undefined) {
      setFieldValue('ids.join_eui', undefined)
      setIsClaiming(undefined)
      setJoinEUIDescription(msg.joinEUIDesc)
    } else {
      const claim = await tts.DeviceClaim.GetInfoByJoinEUI({ join_eui: values.ids.join_eui })
      const supportsClaiming = claim.supports_claiming ?? false
      setIsClaiming(supportsClaiming)
      setJoinEUIDescription(isClaiming ? msg.claiming : msg.registration)
    }
  }, [values, setFieldValue, isClaiming, setIsClaiming])

  const hasTemplate = Boolean(template)
  const hasCompletedRepo = version_ids && Object.values(version_ids).every(value => value)
  const hasCompletedManual
  const showProvisioning =
    (hasCompletedRepo && _inputMethod === 'device-repository') ||
    (hasCompletedManual && _inputMethod === 'manual')

  React.useEffect(() => {
    // Merge the new device template with other form values.
    if (hasTemplate && values._inputMethod === 'device-repository') {
      const { end_device } = template
      setValues(merge(values, end_device), false)
    }
  }, [hasTemplate, template, setValues, values])

  return (
    showProvisioning && (
      <>
        <Form.SubTitle title={m.provisioningTitle} />
        <Form.Field
          title={sharedMessages.joinEUI}
          name="ids.join_eui"
          type="byte"
          min={8}
          max={8}
          required
          component={Input}
          actionDisable={!Boolean(values.ids.join_eui)}
          action={{
            type: 'button',
            disable: !Boolean(values.ids.join_eui),
            title: isClaiming === undefined ? msg.confirm : msg.reset,
            message: isClaiming === undefined ? msg.confirm : msg.reset,
            onClick: isClaimable,
            raw: true,
          }}
          tooltipId={tooltipIds.JOIN_EUI}
          description={joinEUIDescription}
        />
        {isClaiming === true && (
          <DeviceClaimingFormSection
            appId={appId}
            fetchDevEUICounter={fetchDevEUICounter}
            issueDevEUI={issueDevEUI}
          />
        )}
        {isClaiming === false && (
          <DeviceRegistrationFormSection
            appId={appId}
            fetchDevEUICounter={fetchDevEUICounter}
            issueDevEUI={issueDevEUI}
            mayEditKeys={mayEditKeys}
          />
        )}
        {isClaiming !== undefined && (
          <>
            <Form.Field title={m.afterRegistration} name="_registration" component={Radio.Group}>
              <Radio label={m.singleRegistration} value={REGISTRATION_TYPES.SINGLE} />
              <Radio label={m.multipleRegistration} value={REGISTRATION_TYPES.MULTIPLE} />
            </Form.Field>
            <SubmitBar>
              <Form.Submit
                component={SubmitButton}
                message={isClaiming === true ? m.claimEndDevice : m.registerEndDevice}
              />
            </SubmitBar>
          </>
        )}
      </>
    )
  )
}

DeviceProvisioningFormSection.propTypes = {
  appId: PropTypes.string,
  fetchDevEUICounter: PropTypes.func,
  isClaiming: PropTypes.bool,
  issueDevEUI: PropTypes.func,
  mayEditKeys: PropTypes.bool.isRequired,
  setIsClaiming: PropTypes.func,
}

DeviceProvisioningFormSection.defaultProps = {
  appId: undefined,
  fetchDevEUICounter: () => null,
  issueDevEUI: () => null,
  isClaiming: undefined,
  setIsClaiming: () => null,
}

export { DeviceProvisioningFormSection as default, initialValues }
