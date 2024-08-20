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

import React, { useCallback } from 'react'
import PropTypes from 'prop-types'
import { defineMessages } from 'react-intl'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import Button from '@ttn-lw/components/button'

import AccessPointList from '@console/containers/access-point-list'
import NetworkInterfaceAddressesFormFields from '@console/containers/gateway-managed-gateway/shared/network-interface-addresses-form-fields'
import { getValuesNormalized } from '@console/containers/gateway-managed-gateway/shared/utils'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  useDefaultNetworkInterfaceSettings: 'Use default network interface settings',
  uncheckToSetCustomSettings:
    'Uncheck if you need to set custom IP addresses, subnet mask and DNS server',
  accessPointAndSsid: 'Access point / SSID',
  wifiPassword: 'WiFi password',
  ssid: 'SSID',
  isSet: '(is set)',
})

const GatewayWifiProfilesFormFields = ({ namePrefix }) => {
  const { values, setFieldValue, setFieldTouched } = useFormContext()

  const valuesNormalized = getValuesNormalized(namePrefix, values)

  const canTypePassword =
    !valuesNormalized._access_point?.is_password_set ||
    valuesNormalized._access_point?.type === 'other'

  const handleResetPassword = useCallback(() => {
    setFieldValue(`${namePrefix}_access_point.is_password_set`, false)
  }, [namePrefix, setFieldValue])

  const handleChangeAccessPoint = useCallback(
    accessPoint => {
      if (Boolean(accessPoint.ssid) || !accessPoint.is_password_set) {
        setFieldValue(`${namePrefix}ssid`, accessPoint?.ssid)
        setFieldTouched(`${namePrefix}password`, false)
        setFieldValue(`${namePrefix}password`, '')
      }
    },
    [namePrefix, setFieldTouched, setFieldValue],
  )

  return (
    <>
      {valuesNormalized.profile_id !== 'non-shared' && (
        <Form.Field
          title={sharedMessages.profileName}
          name={`${namePrefix}profile_name`}
          component={Input}
          required
        />
      )}

      <Form.Field
        title={m.accessPointAndSsid}
        name={`${namePrefix}_access_point`}
        onChange={handleChangeAccessPoint}
        component={AccessPointList}
        required
        ssid={valuesNormalized.ssid}
      />
      {valuesNormalized._access_point.type === 'other' && (
        <Form.Field title={m.ssid} name={`${namePrefix}ssid`} component={Input} required />
      )}
      {(valuesNormalized._access_point.type === 'other' ||
        (Boolean(valuesNormalized._access_point.authentication_mode) &&
          valuesNormalized._access_point.authentication_mode !== 'open')) && (
        <Form.Field
          title={m.wifiPassword}
          name={`${namePrefix}password`}
          type="password"
          component={Input}
          readOnly={!canTypePassword}
          placeholder={!canTypePassword ? m.isSet : undefined}
          required={valuesNormalized._access_point?.type !== 'other'}
        >
          {!canTypePassword && (
            <Button
              className="ml-cs-xs"
              type="button"
              message={sharedMessages.reset}
              onClick={handleResetPassword}
              secondary
            />
          )}
        </Form.Field>
      )}

      <Form.Field
        name={`${namePrefix}_default_network_interface`}
        component={Checkbox}
        label={m.useDefaultNetworkInterfaceSettings}
        description={m.uncheckToSetCustomSettings}
      />

      {!Boolean(valuesNormalized._default_network_interface) && (
        <NetworkInterfaceAddressesFormFields namePrefix={namePrefix} />
      )}
    </>
  )
}

GatewayWifiProfilesFormFields.propTypes = {
  namePrefix: PropTypes.string,
}

GatewayWifiProfilesFormFields.defaultProps = {
  namePrefix: '',
}

export default GatewayWifiProfilesFormFields
