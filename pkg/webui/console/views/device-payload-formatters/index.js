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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import { Switch, Route, Redirect } from 'react-router'

import sharedMessages from '../../../lib/shared-messages'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import Tab from '../../../components/tabs'
import NotFoundRoute from '../../../lib/components/not-found-route'

import DeviceUplinkPayloadFormatters from '../../containers/device-payload-formatters/uplink'
import DeviceDownlinkPayloadFormatters from '../../containers/device-payload-formatters/downlink'
import {
  selectApplicationIsLinked,
  selectApplicationLink,
  selectApplicationLinkFetching,
  selectSelectedApplicationId,
} from '../../store/selectors/applications'
import { selectSelectedDeviceId } from '../../store/selectors/device'

import style from './device-payload-formatters.styl'

@connect(function(state) {
  const link = selectApplicationLink(state)
  const fetching = selectApplicationLinkFetching(state)

  return {
    appId: selectSelectedApplicationId(state),
    devId: selectSelectedDeviceId(state),
    fetching: fetching || !link,
    linked: selectApplicationIsLinked(state),
  }
})
@withBreadcrumb('device.single.payload-formatters', function(props) {
  const { appId, devId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/payload-formatters`}
      icon="link"
      content={sharedMessages.payloadFormatters}
    />
  )
})
export default class DevicePayloadFormatters extends Component {
  render() {
    const {
      match: { url },
    } = this.props

    const tabs = [
      { title: sharedMessages.uplink, name: 'uplink', link: `${url}/uplink` },
      { title: sharedMessages.downlink, name: 'downlink', link: `${url}/downlink` },
    ]

    return (
      <Container>
        <Row>
          <Col>
            <Tab className={style.tabs} tabs={tabs} divider />
          </Col>
        </Row>
        <Row>
          <Col>
            <Switch>
              <Redirect exact from={url} to={`${url}/uplink`} />
              <Route exact path={`${url}/uplink`} component={DeviceUplinkPayloadFormatters} />
              <Route exact path={`${url}/downlink`} component={DeviceDownlinkPayloadFormatters} />
              <NotFoundRoute />
            </Switch>
          </Col>
        </Row>
      </Container>
    )
  }
}
