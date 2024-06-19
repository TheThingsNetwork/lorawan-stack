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

import React, { useCallback, useState } from 'react'
import { useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'

import {
  CONNECTION_TYPES,
  getFormTypeMessage,
} from '@console/containers/gateway-the-things-station/connection-profiles/utils'
import validationSchema from '@console/containers/gateway-the-things-station/connection-profiles/validation-schema'
import AccessPointList from '@console/containers/access-point-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import m from './messages'

const GatewayConnectionProfilesForm = () => {
  const [error, setError] = useState(undefined)
  const { gtwId, type, profileId } = useParams()

  const isEdit = Boolean(profileId)

  const baseUrl = `/gateways/${gtwId}/the-things-station/connection-profiles/${type}`

  useBreadcrumbs(
    'gtws.single.the-things-station.connection-profiles.form',
    <Breadcrumb
      path={isEdit ? `${baseUrl}/edit/${profileId}` : `${baseUrl}/add`}
      content={getFormTypeMessage(type, profileId)}
    />,
  )

  const handleSubmit = useCallback(values => {
    try {
      console.log(values)
    } catch (e) {
      setError(e)
    }
  }, [])

  const initialValues = {
    _connection_type: type,
    name: '',
    ...(type === CONNECTION_TYPES.WIFI && {
      access_point: {
        _type: 'all',
        ssid: '',
        password: '',
        security: '',
        signal_strength: 0,
        is_active: true,
      },
    }),
    default_network_interface: true,
    ip_address: '',
    subnet_mask: '',
    dns_servers: [''],
  }

  return (
    <>
      <PageTitle title={getFormTypeMessage(type, profileId)} />
      <Form
        error={error}
        onSubmit={handleSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
      >
        {({ values }) => (
          <>
            <Form.Field title={m.profileName} name="name" component={Input} required />
            {values._connection_type === CONNECTION_TYPES.WIFI && (
              <>
                <Form.Field
                  title={m.accessPointAndSsid}
                  name="access_point"
                  component={AccessPointList}
                  required
                />
                {values.access_point._type === 'other' && (
                  <Form.Field title={m.ssid} name="access_point.ssid" component={Input} required />
                )}
                {(values.access_point._type === 'other' ||
                  values.access_point.password.length > 1) && (
                  <Form.Field
                    title={m.wifiPassword}
                    name="access_point.password"
                    type="password"
                    component={Input}
                    required={values.access_point._type !== 'other'}
                  />
                )}
              </>
            )}

            <Form.Field
              name="default_network_interface"
              component={Checkbox}
              label={m.useDefaultNetworkInterfaceSettings}
              description={m.uncheckToSetCustomSettings}
              tooltipId={tooltipIds.DEFAULT_NETWORK_INTERFACE}
            />

            {!Boolean(values.default_network_interface) && (
              <>
                <Form.Field title={m.ipAddress} name="ip_address" component={Input} />
                <Form.Field title={m.subnetMask} name="subnet_mask" component={Input} />
                <Form.Field
                  name="dns_servers"
                  title={m.dnsServers}
                  addMessage={m.addServerAddress}
                  component={KeyValueMap}
                  indexAsKey
                  valuePlaceholder={m.dnsServerPlaceholder}
                />
              </>
            )}

            <SubmitBar>
              <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
            </SubmitBar>
          </>
        )}
      </Form>
    </>
  )
}

export default GatewayConnectionProfilesForm
