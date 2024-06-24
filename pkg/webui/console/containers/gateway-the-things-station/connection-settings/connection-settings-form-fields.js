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
import PropTypes from 'prop-types'
import { defineMessages } from 'react-intl'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Select from '@ttn-lw/components/select'

import Message from '@ttn-lw/lib/components/message'

import { CONNECTION_TYPES } from '@console/containers/gateway-the-things-station/utils'
import GatewayConnectionProfilesFormFields from '@console/containers/gateway-the-things-station/connection-profiles/connection-profiles-form-fields'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  settingsProfile: 'Settings profile',
  profileDescription: 'Connection settings profiles can be shared within the same organization',
  wifiConnection: 'WiFi connection',
  ethernetConnection: 'Ethernet connection',
})

const getTitle = type => {
  switch (type) {
    case CONNECTION_TYPES.WIFI:
      return m.wifiConnection
    default:
      return m.ethernetConnection
  }
}

const GatewayConnectionSettingsFormFields = ({ index }) => {
  const { values } = useFormContext()
  const profileOptions = [
    { value: '0', label: sharedMessages.yourself },
    { value: '1', label: 'TTI' },
    { value: '2', label: 'Create new profile' },
  ]

  return (
    <>
      <Message component="h3" content={getTitle(values.settings[index]._connection_type)} />
      <div className="d-flex al-center gap-cs-m">
        <Form.Field
          name={`settings.${index}.profileFrom`}
          title={sharedMessages.showProfilesOf}
          component={Select}
          options={profileOptions}
          tooltipId={tooltipIds.GATEWAY_SHOW_PROFILES}
          description={m.profileDescription}
        />
        <Form.Field
          name={`settings.${index}.profile`}
          title={m.settingsProfile}
          component={Select}
          options={profileOptions}
          tooltipId={tooltipIds.GATEWAY_SHOW_PROFILES}
        />
      </div>
      {values.settings[index].profile === '2' && (
        <GatewayConnectionProfilesFormFields namePrefix={`settings.${index}.`} />
      )}
    </>
  )
}

GatewayConnectionSettingsFormFields.propTypes = {
  index: PropTypes.number.isRequired,
}

export default GatewayConnectionSettingsFormFields
