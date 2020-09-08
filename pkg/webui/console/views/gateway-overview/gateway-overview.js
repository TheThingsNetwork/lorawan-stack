// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Col, Row } from 'react-grid-system'

import DataSheet from '@ttn-lw/components/data-sheet'
import Tag from '@ttn-lw/components/tag'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import GatewayMap from '@console/components/gateway-map'

import GatewayEvents from '@console/containers/gateway-events'
// Import GatewayConnection from '@console/containers/gateway-connection'
import GatewayTitleSection from '@console/containers/gateway-title-section'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewGatewayInfo } from '@console/lib/feature-checks'

@withFeatureRequirement(mayViewGatewayInfo, {
  redirect: '/',
})
export default class GatewayOverview extends React.Component {
  static propTypes = {
    gateway: PropTypes.gateway.isRequired,
    gtwId: PropTypes.string.isRequired,
  }

  render() {
    const {
      gtwId,
      gateway: {
        ids,
        description,
        created_at,
        updated_at,
        frequency_plan_id,
        gateway_server_address,
      },
    } = this.props

    const sheetData = [
      {
        header: sharedMessages.generalInformation,
        items: [
          {
            key: sharedMessages.gatewayID,
            value: gtwId,
            type: 'code',
            sensitive: false,
          },
          {
            key: sharedMessages.gatewayEUI,
            value: ids.eui,
            type: 'code',
            sensitive: false,
          },
          {
            key: sharedMessages.gatewayDescription,
            value: description || <Message content={sharedMessages.none} />,
          },
          {
            key: sharedMessages.createdAt,
            value: <DateTime value={created_at} />,
          },
          {
            key: sharedMessages.updatedAt,
            value: <DateTime value={updated_at} />,
          },
          {
            key: sharedMessages.gatewayServerAddress,
            value: gateway_server_address,
            type: 'code',
            sensitive: false,
          },
        ],
      },
      {
        header: sharedMessages.lorawanInformation,
        items: [
          {
            key: sharedMessages.frequencyPlan,
            value: frequency_plan_id ? <Tag content={frequency_plan_id} /> : undefined,
          },
        ],
      },
    ]

    return (
      <Container>
        <IntlHelmet title={sharedMessages.overview} />
        <Row>
          <Col sm={12}>
            <GatewayTitleSection gtwId={gtwId} />
          </Col>
        </Row>
        <Row>
          <Col sm={12} lg={6}>
            <DataSheet data={sheetData} />
          </Col>
          <Col sm={12} lg={6}>
            <GatewayEvents gtwId={gtwId} widget />
            <GatewayMap gtwId={gtwId} gateway={this.props.gateway} />
          </Col>
        </Row>
      </Container>
    )
  }
}
