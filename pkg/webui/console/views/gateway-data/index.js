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
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Message from '../../../lib/components/message'
import GatewayEvents from '../../containers/gateway-events'

import { getGatewayId } from '../../../lib/selectors/id'

import style from './gateway-data.styl'

const m = defineMessages({
  gtwData: 'Gateway Data',
})

@connect(function (state) {
  const gateway = state.gateway.gateway

  return {
    gtwId: getGatewayId(gateway),
  }
})
@withBreadcrumb('gateways.single.data', function (props) {
  return (
    <Breadcrumb
      path={`/console/gateways/${props.gtwId}/data`}
      icon="data"
      content={sharedMessages.data}
    />
  )
})
export default class Data extends React.Component {

  render () {
    const { gtwId } = this.props

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={m.gtwData} />
            <Message component="h2" content={m.gtwData} />
          </Col>
        </Row>
        <Row>
          <Col sm={12} className={style.wrapper}>
            <GatewayEvents
              gtwId={gtwId}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
