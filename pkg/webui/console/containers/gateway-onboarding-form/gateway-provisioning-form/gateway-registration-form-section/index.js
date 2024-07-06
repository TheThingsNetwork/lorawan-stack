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

import React, { useEffect } from 'react'
import { useFormikContext } from 'formik'
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import Link from '@ttn-lw/components/link'
import SubmitBar from '@ttn-lw/components/submit-bar'
import FormSubmit from '@ttn-lw/components/form/submit'
import SubmitButton from '@ttn-lw/components/submit-button'

import Message from '@ttn-lw/lib/components/message'

import GsFrequencyPlansSelect from '@console/containers/freq-plans-select/gs-frequency-plan-select'

import { selectGsConfig } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

const { enabled: gsEnabled, base_url: gsBaseURL } = selectGsConfig()

const m = defineMessages({
  requireAuthenticatedConnectionDescription:
    'Select which information can be seen by other network participants, including {packetBrokerURL}',
  shareGatewayInfoDescription:
    'Choose this option eg. if your gateway is powered by {loraBasicStationURL}',
})

const PacketBrokerURL = (
  <Link.Anchor external secondary href="https://packetbroker.net/">
    Packet Broker
  </Link.Anchor>
)

const LoraBasicStationURL = (
  <Link.Anchor
    external
    secondary
    href="https://www.thethingsindustries.com/docs/gateways/lora-basics-station/"
  >
    LoRa Basic Station
  </Link.Anchor>
)

const initialValues = {
  ids: {
    gateway_id: '',
  },
  status_public: true,
  location_public: true,
  frequency_plan_ids: [''],
  name: '',
  require_authenticated_connection: false,
  schedule_anytime_delay: '0.530s',
  gateway_server_address: gsEnabled ? new URL(gsBaseURL).hostname : '',
  _create_api_key_cups: false,
  _create_api_key_lns: false,
}

const GatewayRegistrationFormSections = () => {
  const { values, addToFieldRegistry, removeFromFieldRegistry } = useFormikContext()
  const gatewayEui = values.ids?.eui

  // Register hidden fields so they don't get cleaned.
  useEffect(() => {
    const hiddenField = ['gateway_server_address', 'enforce_duty_cycle', 'schedule_anytime_delay']
    addToFieldRegistry(...hiddenField)
    return () => removeFromFieldRegistry(...hiddenField)
  }, [addToFieldRegistry, removeFromFieldRegistry])

  return (
    <>
      <Form.Field
        title={sharedMessages.gatewayID}
        name="ids.gateway_id"
        placeholder={sharedMessages.gatewayIdPlaceholder}
        component={Input}
        tooltipId={tooltipIds.GATEWAY_ID}
        autoFocus={!Boolean(gatewayEui)}
        required
      />
      <Form.Field
        title={sharedMessages.gatewayName}
        placeholder={sharedMessages.gatewayNamePlaceholder}
        name="name"
        component={Input}
        tooltipId={tooltipIds.GATEWAY_NAME}
        autoFocus={Boolean(gatewayEui)}
      />
      {gsEnabled && <GsFrequencyPlansSelect />}
      <Form.Field
        name="require_authenticated_connection"
        component={Checkbox}
        label={sharedMessages.requireAuthenticatedConnection}
        tooltipId={tooltipIds.REQUIRE_AUTHENTICATED_CONNECTION}
        description={{
          ...m.shareGatewayInfoDescription,
          values: { loraBasicStationURL: LoraBasicStationURL },
        }}
        className="mb-cs-xs mt-ls-xs"
        labelAsTitle
      />
      {values.require_authenticated_connection && (
        <>
          <Form.Field
            name="_create_api_key_cups"
            component={Checkbox}
            label={sharedMessages.generateAPIKeyCups}
            tooltipId={tooltipIds.GATEWAY_GENERATE_API_KEY_CUPS}
            className="mb-0"
            labelAsTitle
          />
          <Form.Field
            name="_create_api_key_lns"
            component={Checkbox}
            label={sharedMessages.generateAPIKeyLNS}
            className="mb-cs-xl"
            tooltipId={tooltipIds.GATEWAY_GENERATE_API_KEY_LNS}
            labelAsTitle
          />
        </>
      )}
      <Message component="h4" content={sharedMessages.shareGatewayInfo} className="mb-0" />
      <Message
        component="p"
        content={m.requireAuthenticatedConnectionDescription}
        values={{ packetBrokerURL: PacketBrokerURL }}
        className="m-0 mb-cs-xs tc-subtle-gray"
      />
      <Form.Field
        name="status_public"
        component={Checkbox}
        label={sharedMessages.gatewayStatusPublic}
        tooltipId={tooltipIds.GATEWAY_STATUS_PUBLIC}
        className="mb-0"
        labelAsTitle
      />
      <Form.Field
        name="location_public"
        component={Checkbox}
        label={sharedMessages.gatewayLocationPublic}
        tooltipId={tooltipIds.GATEWAY_LOCATION_PUBLIC}
        labelAsTitle
      />
      <SubmitBar>
        <FormSubmit component={SubmitButton} message={sharedMessages.registerGateway} />
      </SubmitBar>
    </>
  )
}

export { GatewayRegistrationFormSections as default, initialValues }
