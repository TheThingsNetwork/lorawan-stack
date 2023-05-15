// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route, Redirect } from 'react-router-dom'

import Tabs from '@ttn-lw/components/tabs'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import DownlinkForm from '@console/components/downlink-form'
import UplinkForm from '@console/components/uplink-form'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayWriteTraffic } from '@console/lib/feature-checks'

import style from './device-messaging.styl'

const DeviceMessaging = ({ match, mayScheduleDownlinks, maySendUplink }) => {
  const { url } = match

  const tabs =
    mayScheduleDownlinks && maySendUplink
      ? [
          { title: sharedMessages.uplink, name: 'uplink', link: `${url}/uplink` },
          { title: sharedMessages.downlink, name: 'downlink', link: `${url}/downlink` },
        ]
      : []

  const uplinkPath = `${url}/uplink`
  const downlinkPath = `${url}/downlink`

  return (
    <Container>
      <IntlHelmet title={sharedMessages.messaging} />
      <Row>
        {tabs.length > 0 && (
          <Col sm={12}>
            <Tabs className={style.tabs} tabs={tabs} divider />
          </Col>
        )}
        <Col lg={8} md={12}>
          <Routes>
            <Redirect exact from={url} to={maySendUplink ? uplinkPath : downlinkPath} />
            {maySendUplink && <Route path={uplinkPath} component={UplinkForm} />}
            {mayScheduleDownlinks && <Route path={downlinkPath} component={DownlinkForm} />}
            <NotFoundRoute />
          </Routes>
        </Col>
      </Row>
    </Container>
  )
}

DeviceMessaging.propTypes = {
  match: PropTypes.match.isRequired,
  mayScheduleDownlinks: PropTypes.bool.isRequired,
  maySendUplink: PropTypes.bool.isRequired,
}

export default withFeatureRequirement(mayWriteTraffic, {
  redirect: ({ appId, devId }) => `/applications/${appId}/devices/${devId}`,
})(DeviceMessaging)
