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
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'

import DataSheet from '../../../components/data-sheet'
import GatewayStatistics from '../../containers/gateway-statistics'
import GatewayEvents from '../../containers/gateway-events'
import Tag from '../../../components/tag'

import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'
import DateTime from '../../../lib/components/date-time'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'

import { selectSelectedGateway as gatewaySelector } from '../../store/selectors/gateways'
import { getGatewayId as idSelector } from '../../../lib/selectors/id'

import style from './gateway-overview.styl'

@connect(function(state, props) {
  const gtw = gatewaySelector(state, props)

  return {
    gtwId: idSelector(gtw),
    gateway: gtw,
  }
})
@bind
export default class GatewayOverview extends React.Component {
  static propTypes = {
    gateway: PropTypes.gateway.isRequired,
    gtwId: PropTypes.string.isRequired,
  }

  get gatewayInfo() {
    const { gtwId, gateway } = this.props
    const {
      ids,
      name,
      description,
      created_at,
      updated_at,
      frequency_plan_id,
      gateway_server_address,
    } = gateway

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
            value: <Tag content={frequency_plan_id} />,
          },
        ],
      },
    ]

    return (
      <div className={style.overviewInfo}>
        <h2 className={style.title}>{name || gtwId}</h2>
        <GatewayStatistics className={style.statistics} gtwId={gtwId} />
        <DataSheet data={sheetData} />
      </div>
    )
  }

  render() {
    const { gtwId } = this.props

    return (
      <Container>
        <IntlHelmet title={sharedMessages.overview} />
        <Row>
          <Col sm={12} lg={6}>
            {this.gatewayInfo}
          </Col>
          <Col sm={12} lg={6}>
            <div className={style.latestEvents}>
              <GatewayEvents gtwId={gtwId} widget />
            </div>
          </Col>
        </Row>
      </Container>
    )
  }
}
