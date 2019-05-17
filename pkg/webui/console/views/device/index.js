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
import { Switch, Route } from 'react-router'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import Spinner from '../../../components/spinner'
import Tabs from '../../../components/tabs'
import IntlHelmet from '../../../lib/components/intl-helmet'

import DeviceOverview from '../device-overview'
import DeviceData from '../device-data'
import DeviceGeneralSettings from '../device-general-settings'
import DeviceLocation from '../device-location'

import {
  getDevice,
  stopDeviceEventsStream,
} from '../../store/actions/device'

import style from './device.styl'

const m = defineMessages({
  title: '%s - {deviceName} - The Things Network Console',
})

@connect(function ({ device, application }, props) {
  return {
    appName: application.application.name,
    deviceName: device.device && device.device.name,
    devIds: device.device && device.device.ids,
    devId: props.match.params.devId,
    fetching: device.fetching,
    error: device.error,
  }
}, dispatch => ({
  getDevice: (appId, devId, selectors, config) =>
    dispatch(getDevice(appId, devId, selectors, config)),
  stopStream: id => dispatch(stopDeviceEventsStream(id)),
}))
@withBreadcrumb('device.single', function (props) {
  const { devId } = props
  const { appId } = props.match.params
  return (
    <Breadcrumb
      path={`/console/applications/${appId}/devices/${devId}`}
      icon="device"
      content={devId}
    />
  )
})
export default class Device extends React.Component {

  componentDidMount () {
    const { getDevice, devId, match } = this.props
    const { appId } = match.params

    getDevice(
      appId,
      devId,
      [
        'name',
        'description',
        'session',
        'version_ids',
        'root_keys',
        'frequency_plan_id',
        'mac_settings.resets_f_cnt',
        'resets_join_nonces',
        'supports_class_c',
        'supports_join',
        'lorawan_version',
        'lorawan_phy_version',
        'locations',
      ],
      { ignoreNotFound: true })
  }

  componentWillUnmount () {
    const { devIds, stopStream } = this.props

    stopStream(devIds)
  }


  render () {
    const { fetching, error, match, devId, deviceName } = this.props
    const { appId } = match.params

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    // show any device fetching error, e.g. not found, no rights, etc
    if (error) {
      throw error
    }

    const basePath = `/console/applications/${appId}/devices/${devId}`

    const tabs = [
      { title: sharedMessages.overview, name: 'overview', link: basePath },
      { title: sharedMessages.data, name: 'data', link: `${basePath}/data` },
      { title: sharedMessages.location, name: 'location', link: `${basePath}/location` },
      { title: sharedMessages.payloadFormats, name: 'develop' },
      { title: sharedMessages.generalSettings, name: 'general-settings', link: `${basePath}/general-settings` },
    ]

    return (
      <React.Fragment>
        <IntlHelmet
          titleTemplate={m.title} values={{ deviceName: deviceName || devId }}
        />
        <Container>
          <Row>
            <Col lg={12}>
              <h2 className={style.title}>{deviceName || devId}</h2>
              <Tabs
                narrow
                tabs={tabs}
                className={style.tabs}
              />
            </Col>
          </Row>
        </Container>
        <Switch>
          <Route exact path={basePath} component={DeviceOverview} />
          <Route exact path={`${basePath}/data`} component={DeviceData} />
          <Route exact path={`${basePath}/location`} component={DeviceLocation} />
          <Route exact path={`${basePath}/general-settings`} component={DeviceGeneralSettings} />
        </Switch>
      </React.Fragment>
    )
  }
}
