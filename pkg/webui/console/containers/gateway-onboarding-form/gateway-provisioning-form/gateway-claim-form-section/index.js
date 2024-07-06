// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useEffect } from 'react'
import { defineMessages } from 'react-intl'
import { useFormikContext } from 'formik'

import frequencyPlans from '@console/constants/frequency-plans'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Notification from '@ttn-lw/components/notification'
import SubmitBar from '@ttn-lw/components/submit-bar'
import FormSubmit from '@ttn-lw/components/form/submit'
import SubmitButton from '@ttn-lw/components/submit-button'
import Link from '@ttn-lw/components/link'

import { GsFrequencyPlansSelect as FrequencyPlansSelect } from '@console/containers/freq-plans-select'

import { selectGsConfig } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

const { enabled: gsEnabled, base_url: gsBaseURL } = selectGsConfig()

const m = defineMessages({
  claimWarning:
    'We detected that your gateway is a <strong>The Things Station</strong>. To claim this gateway, please use the claim authentication code printed on the <link1>inside of the mounting lid</link1> or <link2>scan the QR code</link2> to claim instantly.',
  preselectCommonOption:
    'We preselect a common option for the region you are in. Please verify the gateway operates on the selected frequency plan.',
})

const getCupsUri = gsUrl => {
  try {
    const gtwUrl = new URL(gsUrl)

    return `${gtwUrl.protocol}//${gtwUrl.hostname}:443`
  } catch (err) {
    // Ignore any errors related to the URL composition.
  }

  return ''
}

const initialValues = {
  authenticated_identifiers: {
    authentication_code: '',
    gateway_eui: '',
  },
  target_gateway_id: '',
  target_frequency_plan_id: '',
  target_gateway_server_address: gsEnabled ? getHostFromUrl(gsBaseURL) : '',
  cups_redirection: {
    target_cups_uri: gsEnabled ? getCupsUri(gsBaseURL) : '',
  },
}

// This is the TrackNet prefix that all TTIGs use.
// const TRACKNET_PREFIX = '58A0CBFFFE'

const GatewayClaimFormSection = () => {
  const { values, addToFieldRegistry, removeFromFieldRegistry } = useFormikContext()

  const isTTIG = values?._inputMethod === 'tts'

  // Register hidden fields so they don't get cleaned.
  useEffect(() => {
    const hiddenFields = ['target_gateway_server_address', 'cups_redirection']
    addToFieldRegistry(...hiddenFields)
    return () => removeFromFieldRegistry(...hiddenFields)
  }, [addToFieldRegistry, removeFromFieldRegistry])

  return (
    <>
      {isTTIG && (
        <Form.InfoField>
          <Notification
            small
            info
            content={m.claimWarning}
            className="mb-0"
            messageValues={{
              strong: msg => <strong>{msg}</strong>,
              link1: txt => (
                <Link to={`#`} secondary>
                  {txt}
                </Link>
              ),
              link2: txt => (
                <Link to={`#`} secondary>
                  {txt}
                </Link>
              ),
            }}
          />
        </Form.InfoField>
      )}
      <Form.Field
        required
        title={sharedMessages.claimAuthCode}
        name="authenticated_identifiers.authentication_code"
        tooltipId={tooltipIds.CLAIM_AUTH_CODE}
        component={Input}
        encode={btoa}
        decode={atob}
        autoFocus
      />
      <Form.Field
        title={sharedMessages.gatewayID}
        name="target_gateway_id"
        placeholder={sharedMessages.gatewayIdPlaceholder}
        required
        component={Input}
        tooltipId={tooltipIds.GATEWAY_ID}
      />

      {gsEnabled && (
        <FrequencyPlansSelect
          name="target_frequency_plan_id"
          menuPlacement="top"
          tooltipId={tooltipIds.FREQUENCY_PLAN}
          warning={
            values.frequency_plan_id === frequencyPlans.EMPTY_FREQ_PLAN
              ? sharedMessages.frequencyPlanWarning
              : undefined
          }
          description={isTTIG ? m.preselectCommonOption : undefined}
          required
        />
      )}
      <SubmitBar>
        <FormSubmit component={SubmitButton} message={sharedMessages.claimGateway} />
      </SubmitBar>
    </>
  )
}

export { GatewayClaimFormSection as default, initialValues }
