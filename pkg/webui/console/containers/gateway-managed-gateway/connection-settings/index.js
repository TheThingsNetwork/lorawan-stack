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

import React, { useCallback, useMemo, useState } from 'react'
import { defineMessages } from 'react-intl'
import { useParams, useSearchParams } from 'react-router-dom'
import { Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Notification from '@ttn-lw/components/notification'
import Form from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import validationSchema from '@console/containers/gateway-managed-gateway/connection-settings/validation-schema'
import {
  CONNECTION_TYPES,
  getInitialWifiProfile,
  initialEthernetProfile,
} from '@console/containers/gateway-managed-gateway/utils'
import GatewayConnectionSettingsFormFields from '@console/containers/gateway-managed-gateway/connection-settings/connection-settings-form-fields'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  firstNotification:
    'You have just claimed a managed gateway. To connect it to WiFi or ethernet you can configure those connections here. The preprovisioned cellular backhaul typically connects automatically.',
})

const GatewayConnectionSettings = () => {
  const { gtwId } = useParams()
  const [searchParams] = useSearchParams()
  const isFirstClaim = Boolean(searchParams.get('claimed'))
  const [error, setError] = useState(undefined)

  useBreadcrumbs(
    'gtws.single.managed-gateway.connection-settings',
    <Breadcrumb
      path={`/gateways/${gtwId}/managed-gateway/connection-settings`}
      content={sharedMessages.connectionSettings}
    />,
  )

  const handleSubmit = useCallback(values => {
    try {
      console.log(values)
    } catch (e) {
      setError(e)
    }
  }, [])

  const initialValues = useMemo(
    () => ({
      settings: [
        {
          _connection_type: CONNECTION_TYPES.WIFI,
          profile: '',
          ...getInitialWifiProfile(false),
        },
        {
          _connection_type: CONNECTION_TYPES.ETHERNET,
          ...initialEthernetProfile,
        },
      ],
    }),
    [],
  )

  return (
    <>
      <PageTitle title={sharedMessages.connectionSettings} />
      <Row>
        <Col lg={8} md={6} sm={12}>
          {isFirstClaim && <Notification info small content={m.firstNotification} />}
          <Form
            error={error}
            onSubmit={handleSubmit}
            initialValues={initialValues}
            validationSchema={validationSchema}
          >
            <>
              <GatewayConnectionSettingsFormFields />

              <SubmitBar>
                <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
              </SubmitBar>
            </>
          </Form>
        </Col>
        <Col lg={4} md={6} sm={12}>
          <div />
        </Col>
      </Row>
    </>
  )
}

export default GatewayConnectionSettings
