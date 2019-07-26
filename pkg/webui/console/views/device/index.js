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
import DevicePayloadFormatters from '../device-payload-formatters'

import {
  getDevice,
  stopDeviceEventsStream,
} from '../../store/actions/device'

import withEnv, { EnvProvider } from '../../../lib/components/env'
import { selectDeviceFetching, selectGetDeviceError } from '../../store/selectors/device'

import style from './device.styl'


@connect(function (state, props) {
  const { device } = state
  return {
    device: device.device,
    deviceName: device.device && device.device.name,
    devIds: device.device && device.device.ids,
    devId: props.match.params.devId,
    fetching: selectDeviceFetching(state),
    error: selectGetDeviceError(state),
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

@withEnv
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
        'network_server_address',
        'application_server_address',
        'join_server_address',
        'locations',
        'formatters',
      ],
      { ignoreNotFound: true })
  }

  componentWillUnmount () {
    const { devIds, stopStream } = this.props

    stopStream(devIds)
  }

  render () {
    const {
      fetching,
      error,
      location: { pathname },
      match: { params: { appId }},
      devId,
      deviceName,
      device,
      env,
    } = this.props

    if (error) {
      throw error
    }

    if (fetching || !device) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    const basePath = `/console/applications/${appId}/devices/${devId}`

    // Prevent default redirect to uplink when tab is already open
    const payloadFormattersLink =
    pathname.startsWith(`${basePath}/payload-formatters`)
      ? pathname : `${basePath}/payload-formatters`

    const tabs = [
      { title: sharedMessages.overview, name: 'overview', link: basePath },
      { title: sharedMessages.data, name: 'data', link: `${basePath}/data` },
      { title: sharedMessages.location, name: 'location', link: `${basePath}/location` },
      { title: sharedMessages.payloadFormatters, name: 'develop', link: payloadFormattersLink, exact: false },
      { title: sharedMessages.generalSettings, name: 'general-settings', link: `${basePath}/general-settings` },
    ]

    return (
      <EnvProvider env={env}>
        <IntlHelmet
          titleTemplate={`%s - ${deviceName || devId} - ${env.site_name}`}
        />
        <Container>
          <Row>
            <Col lg={12}>
              <h2 className={style.title}>{deviceName || devId}</h2>
              <Tabs
                className={style.tabs}
                narrow
                tabs={tabs}
              />
            </Col>
          </Row>
        </Container>
        <hr className={style.rule} />
        <Switch>
          <Route exact path={basePath} component={DeviceOverview} />
          <Route exact path={`${basePath}/data`} component={DeviceData} />
          <Route exact path={`${basePath}/location`} component={DeviceLocation} />
          <Route exact path={`${basePath}/general-settings`} component={DeviceGeneralSettings} />
          <Route path={`${basePath}/payload-formatters`} component={DevicePayloadFormatters} />
        </Switch>
      </EnvProvider>
    )
  }
}
