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

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import GatewayEvents from '@console/containers/gateway-events'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewGatewayEvents } from '@console/lib/feature-checks'

import { selectSelectedGatewayId } from '@console/store/selectors/gateways'

import style from './gateway-data.styl'

const m = defineMessages({
  gtwData: 'Gateway data',
})

@connect(state => ({ gtwId: selectSelectedGatewayId(state) }))
@withFeatureRequirement(mayViewGatewayEvents, {
  redirect: ({ gtwId }) => `/gateways/${gtwId}`,
})
@withBreadcrumb('gateways.single.data', ({ gtwId }) => (
  <Breadcrumb path={`/gateways/${gtwId}/data`} content={sharedMessages.data} />
))
export default class Data extends React.Component {
  static propTypes = {
    gtwId: PropTypes.string.isRequired,
  }

  render() {
    const { gtwId } = this.props

    return (
      <Container>
        <PageTitle hideHeading title={m.gtwData} />
        <Row>
          <Col className={style.wrapper}>
            <GatewayEvents gtwId={gtwId} />
          </Col>
        </Row>
      </Container>
    )
  }
}
