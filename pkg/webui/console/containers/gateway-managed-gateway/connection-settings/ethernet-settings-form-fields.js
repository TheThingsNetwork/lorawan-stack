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

import React from 'react'
import { defineMessages } from 'react-intl'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'

import Message from '@ttn-lw/lib/components/message'

import NetworkInterfaceAddressesFormFields from '@console/containers/gateway-managed-gateway/shared/network-interface-addresses-form-fields'

const m = defineMessages({
  ethernetConnection: 'Ethernet connection',
  enableEthernetConnection: 'Enable ethernet connection',
  useStaticIp: 'Use a static IP address',
})

const EthernetSettingsFormFields = () => {
  const { values } = useFormContext()

  return (
    <>
      <Message component="h3" content={m.ethernetConnection} />

      <Form.Field
        name={`ethernet_profile._use_static_ip`}
        component={Checkbox}
        label={m.useStaticIp}
      />
      <NetworkInterfaceAddressesFormFields
        namePrefix={`ethernet_profile.`}
        showOnlyDns={!values.ethernet_profile._use_static_ip}
      />
    </>
  )
}

export default EthernetSettingsFormFields
