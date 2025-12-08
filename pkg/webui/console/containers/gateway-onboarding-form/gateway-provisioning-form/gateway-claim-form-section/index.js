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

import React, { useCallback, useEffect } from 'react'
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
import Tabs from '@ttn-lw/components/tabs'

import Message from '@ttn-lw/lib/components/message'

import { GsFrequencyPlansSelect as FrequencyPlansSelect } from '@console/containers/freq-plans-select'

import { selectGsConfig } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

const { enabled: gsEnabled, base_url: gsBaseURL } = selectGsConfig()
const smUrl = 'https://accounts.thethingsindustries.com'

const m = defineMessages({
  claimWarning:
    'We detected a Managed gateway. To claim this gateway, please use the <strong>owner token printed on the gateway</strong>, or the <strong>gateway fleet owner token</strong>.',
  fleet: 'Fleet',
})

const initialValues = {
  authenticated_identifiers: {
    authentication_code: '',
    gateway_eui: '',
  },
  target_gateway_id: '',
  target_frequency_plan_id: '',
  target_gateway_server_address: gsEnabled ? getHostFromUrl(gsBaseURL) : '',
}

const ownerTokenTypes = [
  { name: 'gateway', title: sharedMessages.gateway },
  { name: 'fleet', title: m.fleet },
]

const GatewayClaimFormSection = () => {
  const { values, addToFieldRegistry, removeFromFieldRegistry, setValues } = useFormikContext()
  const isManaged = values._inputMethod === 'managed'
  const isFleet = values._isFleet

  const [activeOwnerTokenType, setActiveOwnerTokenType] = React.useState(
    isFleet ? 'fleet' : 'gateway',
  )

  const onOwnerTokenTypeChange = useCallback(
    value => {
      setActiveOwnerTokenType(value)
      setValues(values => ({
        ...values,
        authenticated_identifiers: {
          ...values.authenticated_identifiers,
          authentication_code:
            value === 'fleet' && values._fleet_owner_token
              ? btoa(values._fleet_owner_token)
              : value === 'gateway' && values._gtw_owner_token
                ? btoa(values._gtw_owner_token)
                : '',
        },
      }))
    },
    [setValues],
  )

  // Register hidden fields so they don't get cleaned.
  useEffect(() => {
    const hiddenFields = ['target_gateway_server_address']
    addToFieldRegistry(...hiddenFields)
    return () => removeFromFieldRegistry(...hiddenFields)
  }, [addToFieldRegistry, removeFromFieldRegistry])

  return (
    <>
      {isManaged && (
        <>
          <Form.InfoField>
            <Notification
              small
              info
              content={m.claimWarning}
              messageValues={{
                strong: txt => <strong>{txt}</strong>,
              }}
              className="mb-0"
            />
          </Form.InfoField>
          <Message content={sharedMessages.ownerToken} className="fw-bold" />
          <Tabs
            active={activeOwnerTokenType}
            tabs={ownerTokenTypes}
            onTabChange={onOwnerTokenTypeChange}
            toggleStyle
            small
            className="w-content p-0 mb-cs-xs mt-cs-xxs border-none br-m gap-0 fs-s"
          />
          {activeOwnerTokenType === 'fleet' && (
            <Message
              content={sharedMessages.fleetInfo}
              values={{
                Link: val => (
                  <Link.DocLink
                    secondary
                    path="/hardware/gateways/models/thethingsindoorgatewaypro/#subscription"
                  >
                    {val}
                  </Link.DocLink>
                ),
              }}
              className="mb-cs-xs c-text-neutral-light"
              component="div"
            />
          )}
        </>
      )}
      <Form.Field
        required
        title={sharedMessages.ownerToken}
        showTitle={!isManaged}
        name="authenticated_identifiers.authentication_code"
        tooltipId={tooltipIds.CLAIM_AUTH_CODE}
        component={Input}
        description={
          isManaged && activeOwnerTokenType === 'fleet' ? sharedMessages.fleetTokenInfo : undefined
        }
        descriptionValues={{
          Link: val => (
            <Link.Anchor secondary href={`${smUrl}/dashboard`} external>
              {val}
            </Link.Anchor>
          ),
        }}
        encode={btoa}
        decode={atob}
        sensitive
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
