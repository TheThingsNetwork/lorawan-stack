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
import { Switch, Route } from 'react-router-dom'
import { Col, Row, Container } from 'react-grid-system'

import CONNECTION_STATUS from '@console/constants/connection-status'

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Tabs from '@ttn-lw/components/tabs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import withRequest from '@ttn-lw/lib/components/with-request'
import withEnv from '@ttn-lw/lib/components/env'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import DeviceTitleSection from '@console/containers/device-title-section'

import DeviceData from '@console/views/device-data'
import DeviceGeneralSettings from '@console/views/device-general-settings'
import DeviceMessaging from '@console/views/device-messaging'
import DeviceLocation from '@console/views/device-location'
import DevicePayloadFormatters from '@console/views/device-payload-formatters'
import DeviceClaimAuthenticationCode from '@console/views/device-claim-authentication-code'
import DeviceOverview from '@console/views/device-overview'

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import PropTypes from '@ttn-lw/lib/prop-types'
import { selectJsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'
import { combineDeviceIds } from '@ttn-lw/lib/selectors/id'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayReadApplicationDeviceKeys,
  mayScheduleDownlinks,
  maySendUplink,
  mayViewApplicationLink,
  checkFromState,
} from '@console/lib/feature-checks'

import { getDevice, stopDeviceEventsStream } from '@console/store/actions/devices'
import { getApplicationLink } from '@console/store/actions/link'

import {
  selectSelectedDevice,
  selectDeviceFetching,
  selectDeviceError,
  selectDeviceEventsStatus,
} from '@console/store/selectors/devices'
import {
  selectApplicationLinkFetching,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

import style from './device.styl'

@connect(
  (state, props) => {
    const devId = props.match.params.devId
    const appId = selectSelectedApplicationId(state)
    const device = selectSelectedDevice(state)
    const eventsInitialized =
      selectDeviceEventsStatus(state, combineDeviceIds(appId, devId)) !== CONNECTION_STATUS.UNKNOWN

    const fetching =
      selectDeviceFetching(state) ||
      selectApplicationLinkFetching(state) ||
      !eventsInitialized ||
      !Boolean(device)

    return {
      devId,
      appId,
      device,
      mayReadKeys: checkFromState(mayReadApplicationDeviceKeys, state),
      mayScheduleDownlinks: checkFromState(mayScheduleDownlinks, state),
      maySendUplink: checkFromState(maySendUplink, state),
      mayViewLink: checkFromState(mayViewApplicationLink, state),
      fetching,
      error: selectDeviceError(state),
    }
  },
  dispatch => ({
    loadDeviceData: (appId, devId, deviceSelector, linkSelector, mayViewLink) => {
      dispatch(getDevice(appId, devId, deviceSelector, { ignoreNotFound: true }))

      if (mayViewLink) {
        dispatch(getApplicationLink(appId, linkSelector))
      }
    },
    stopStream: id => dispatch(stopDeviceEventsStream(id)),
  }),
)
@withRequest(({ appId, devId, loadDeviceData, mayReadKeys, mayViewLink }) => {
  const linkSelector = ['skip_payload_crypto', 'default_formatters']
  const deviceSelector = [
    'name',
    'description',
    'version_ids',
    'frequency_plan_id',
    'mac_settings',
    'resets_join_nonces',
    'supports_class_b',
    'supports_class_c',
    'supports_join',
    'lorawan_version',
    'lorawan_phy_version',
    'network_server_address',
    'application_server_address',
    'join_server_address',
    'locations',
    'formatters',
    'multicast',
    'net_id',
    'application_server_id',
    'application_server_kek_label',
    'network_server_kek_label',
    'claim_authentication_code',
    'mac_state.recent_uplinks',
    'pending_mac_state.recent_uplinks',
    'attributes',
    'skip_payload_crypto_override',
  ]

  if (mayReadKeys) {
    deviceSelector.push('session')
    deviceSelector.push('pending_session')
    deviceSelector.push('root_keys')
  }

  return loadDeviceData(appId, devId, deviceSelector, linkSelector, mayViewLink)
})
@withBreadcrumb('device.single', props => {
  const {
    devId,
    appId,
    device: { name },
  } = props
  return <Breadcrumb path={`/applications/${appId}/devices/${devId}`} content={name || devId} />
})
@withEnv
export default class Device extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    devId: PropTypes.string.isRequired,
    device: PropTypes.device.isRequired,
    env: PropTypes.env,
    location: PropTypes.location.isRequired,
    mayScheduleDownlinks: PropTypes.bool.isRequired,
    maySendUplink: PropTypes.bool.isRequired,
    stopStream: PropTypes.func.isRequired,
  }

  static defaultProps = {
    env: undefined,
  }

  componentWillUnmount() {
    const { device, stopStream } = this.props

    stopStream(device.ids)
  }

  render() {
    const {
      location: { pathname },
      appId,
      devId,
      device,
      env: { siteName },
      mayScheduleDownlinks,
      maySendUplink,
    } = this.props
    const { name, join_server_address, supports_join, root_keys, application_server_address } =
      device

    const jsConfig = selectJsConfig()
    const hasJs =
      jsConfig.enabled &&
      join_server_address === getHostnameFromUrl(jsConfig.base_url) &&
      supports_join &&
      Boolean(root_keys)

    const asConfig = selectAsConfig()
    const hasAs =
      asConfig.enabled && application_server_address === getHostnameFromUrl(asConfig.base_url)
    const hideMessaging = !hasAs || !(mayScheduleDownlinks || maySendUplink)
    const hidePayloadFormatters = !hasAs
    const hideClaiming = !hasJs

    const basePath = `/applications/${appId}/devices/${devId}`

    // Prevent default redirect to uplink when tab is already open.
    const payloadFormattersLink = pathname.startsWith(`${basePath}/payload-formatters`)
      ? pathname
      : `${basePath}/payload-formatters`
    const messagingLink = pathname.startsWith(`${basePath}/messaging`)
      ? pathname
      : `${basePath}/messaging`

    const tabs = [
      { title: sharedMessages.overview, name: 'overview', link: basePath },
      { title: sharedMessages.liveData, name: 'data', link: `${basePath}/data` },
      {
        title: sharedMessages.messaging,
        name: 'messaging',
        exact: false,
        link: messagingLink,
        hidden: hideMessaging,
      },
      { title: sharedMessages.location, name: 'location', link: `${basePath}/location` },
      {
        title: sharedMessages.payloadFormatters,
        name: 'develop',
        link: payloadFormattersLink,
        exact: false,
        hidden: hidePayloadFormatters,
      },
      {
        title: sharedMessages.claiming,
        name: 'claim-auth-code',
        link: `${basePath}/claim-auth-code`,
        hidden: hideClaiming,
      },
      {
        title: sharedMessages.generalSettings,
        name: 'general-settings',
        link: `${basePath}/general-settings`,
      },
    ]

    return (
      <React.Fragment>
        <IntlHelmet titleTemplate={`%s - ${name || devId} - ${siteName}`} />
        <div className={style.titleSection}>
          <Container>
            <Row>
              <Col sm={12}>
                <DeviceTitleSection appId={appId} devId={devId}>
                  <Tabs className={style.tabs} narrow tabs={tabs} />
                </DeviceTitleSection>
              </Col>
            </Row>
          </Container>
        </div>
        <Switch>
          <Route exact path={basePath} component={DeviceOverview} />
          <Route exact path={`${basePath}/data`} component={DeviceData} />
          {!hideMessaging && <Route path={`${basePath}/messaging`} component={DeviceMessaging} />}
          <Route exact path={`${basePath}/location`} component={DeviceLocation} />
          <Route exact path={`${basePath}/general-settings`} component={DeviceGeneralSettings} />
          {!hidePayloadFormatters && (
            <Route path={`${basePath}/payload-formatters`} component={DevicePayloadFormatters} />
          )}
          {!hideClaiming && (
            <Route path={`${basePath}/claim-auth-code`} component={DeviceClaimAuthenticationCode} />
          )}
          <NotFoundRoute />
        </Switch>
      </React.Fragment>
    )
  }
}
