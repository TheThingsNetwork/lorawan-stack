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

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import PropTypes from 'prop-types'
import { parseInt } from 'lodash'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Button from '@ttn-lw/components/button'

import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/utils'
import AccessPointList from '@console/containers/access-point-list'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from './messages'

const GatewayConnectionProfilesFormFields = ({ isEdit, namePrefix }) => {
  const { values } = useFormContext()
  const [resetPassword, setResetPassword] = useState(false)

  const valuesNormalized = useMemo(() => {
    if (!namePrefix) return values
    const nameSplitted = namePrefix.split('.')

    return values[nameSplitted[0]][parseInt(nameSplitted[1])]
  }, [namePrefix, values])

  const canTypePassword =
    !isEdit || (isEdit && resetPassword) || valuesNormalized.access_point?._type === 'other'

  const handleRestPassword = useCallback(() => {
    setResetPassword(true)
  }, [])

  useEffect(() => {
    setResetPassword(false)
  }, [valuesNormalized.access_point?.ssid])

  return (
    <>
      <Form.Field
        title={m.profileName}
        name={`${namePrefix}profile_name`}
        component={Input}
        required
      />
      {valuesNormalized._connection_type === CONNECTION_TYPES.WIFI && (
        <>
          <Form.Field
            title={m.accessPointAndSsid}
            name={`${namePrefix}access_point`}
            component={AccessPointList}
            required
          />
          {valuesNormalized.access_point._type === 'other' && (
            <Form.Field
              title={m.ssid}
              name={`${namePrefix}access_point.ssid`}
              component={Input}
              required
            />
          )}
          {(valuesNormalized.access_point._type === 'other' ||
            valuesNormalized.access_point.security === 'WPA2') && (
            <Form.Field
              title={m.wifiPassword}
              name={`${namePrefix}access_point.password`}
              type="password"
              component={Input}
              readOnly={!canTypePassword}
              placeholder={!canTypePassword ? m.isSet : undefined}
              value={!canTypePassword ? undefined : valuesNormalized.access_point?.password}
              required={valuesNormalized.access_point?._type !== 'other'}
            >
              {!canTypePassword && (
                <Button
                  className="ml-cs-xs"
                  type="button"
                  message={sharedMessages.reset}
                  icon="delete"
                  onClick={handleRestPassword}
                />
              )}
            </Form.Field>
          )}
        </>
      )}

      <Form.Field
        name={`${namePrefix}default_network_interface`}
        component={Checkbox}
        label={m.useDefaultNetworkInterfaceSettings}
        description={m.uncheckToSetCustomSettings}
        tooltipId={tooltipIds.DEFAULT_NETWORK_INTERFACE}
      />

      {!Boolean(valuesNormalized.default_network_interface) && (
        <>
          <Form.Field
            name={`${namePrefix}network_interface_addresses.ip_addresses`}
            title={m.ipAddresses}
            addMessage={m.addIpAddress}
            component={KeyValueMap}
            indexAsKey
            valuePlaceholder={m.ipAddressPlaceholder}
            required
          />
          <Form.Field
            title={m.subnetMask}
            name={`${namePrefix}network_interface_addresses.subnet_mask`}
            component={Input}
            required
          />
          <Form.Field
            title={sharedMessages.gateway}
            name={`${namePrefix}network_interface_addresses.gateway`}
            component={Input}
            required
          />
          <Form.Field
            name={`${namePrefix}network_interface_addresses.dns_servers`}
            title={m.dnsServers}
            addMessage={m.addServerAddress}
            component={KeyValueMap}
            indexAsKey
            valuePlaceholder={m.ipAddressPlaceholder}
            required
          />
        </>
      )}
    </>
  )
}

GatewayConnectionProfilesFormFields.propTypes = {
  isEdit: PropTypes.bool,
  namePrefix: PropTypes.string,
}

GatewayConnectionProfilesFormFields.defaultProps = {
  isEdit: false,
  namePrefix: '',
}

export default GatewayConnectionProfilesFormFields
