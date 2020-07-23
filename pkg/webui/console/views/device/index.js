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

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Tabs from '@ttn-lw/components/tabs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import withRequest from '@ttn-lw/lib/components/with-request'
import withEnv from '@ttn-lw/lib/components/env'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import EntityTitleSection from '@console/components/entity-title-section'

import DeviceData from '@console/views/device-data'
import DeviceGeneralSettings from '@console/views/device-general-settings'
import DeviceMessages from '@console/views/device-messages'
import DeviceLocation from '@console/views/device-location'
import DevicePayloadFormatters from '@console/views/device-payload-formatters'
import DeviceClaimAuthenticationCode from '@console/views/device-claim-authentication-code'
import DeviceOverview from '@console/views/device-overview'

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import PropTypes from '@ttn-lw/lib/prop-types'
import { selectJsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayReadApplicationDeviceKeys,
  mayScheduleDownlinks,
  checkFromState,
} from '@console/lib/feature-checks'

import { getDevice, stopDeviceEventsStream } from '@console/store/actions/devices'

import {
  selectSelectedDevice,
  selectDeviceFetching,
  selectDeviceError,
  selectDeviceUplinkFrameCount,
  selectDeviceLastSeen,
} from '@console/store/selectors/devices'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './device.styl'

@connect(
  function(state, props) {
    const devId = props.match.params.devId
    const appId = selectSelectedApplicationId(state)
    const device = selectSelectedDevice(state)

    return {
      devId,
      appId,
      device,
      deviceUplinkFrameCount: selectDeviceUplinkFrameCount(state, appId, devId),
      deviceLastSeen: selectDeviceLastSeen(state, appId, devId),
      mayReadKeys: checkFromState(mayReadApplicationDeviceKeys, state),
      mayScheduleDownlinks: checkFromState(mayScheduleDownlinks, state),
      fetching: selectDeviceFetching(state),
      error: selectDeviceError(state),
    }
  },
  dispatch => ({
    getDevice: (appId, devId, selectors, config) =>
      dispatch(getDevice(appId, devId, selectors, config)),
    stopStream: id => dispatch(stopDeviceEventsStream(id)),
  }),
)
@withRequest(
  ({ appId, devId, getDevice, mayReadKeys }) => {
    const selector = [
      'name',
      'description',
      'version_ids',
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
      'multicast',
      'net_id',
      'application_server_id',
      'application_server_kek_label',
      'network_server_kek_label',
      'claim_authentication_code',
      'recent_uplinks',
      'recent_downlinks',
      'attributes',
      'skip_payload_crypto',
    ]

    if (mayReadKeys) {
      selector.push('session')
      selector.push('root_keys')
    }

    return getDevice(appId, devId, selector, { ignoreNotFound: true })
  },
  ({ fetching, device }) => fetching || !Boolean(device),
)
@withBreadcrumb('device.single', function(props) {
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
    devId: PropTypes.string.isRequired,
    device: PropTypes.device.isRequired,
    deviceLastSeen: PropTypes.string,
    deviceUplinkFrameCount: PropTypes.number,
    env: PropTypes.env,
    location: PropTypes.location.isRequired,
    match: PropTypes.match.isRequired,
    mayScheduleDownlinks: PropTypes.bool.isRequired,
    stopStream: PropTypes.func.isRequired,
  }

  static defaultProps = {
    deviceLastSeen: undefined,
    deviceUplinkFrameCount: undefined,
    env: undefined,
  }

  componentWillUnmount() {
    const { device, stopStream } = this.props

    stopStream(device.ids)
  }

  render() {
    const {
      location: { pathname },
      match: {
        params: { appId },
      },
      devId,
      device: {
        name,
        description,
        join_server_address,
        supports_join,
        root_keys,
        application_server_address,
      },
      deviceUplinkFrameCount,
      deviceLastSeen,
      env: { siteName },
      mayScheduleDownlinks,
    } = this.props

    const jsConfig = selectJsConfig()
    const hasJs =
      jsConfig.enabled &&
      join_server_address === getHostnameFromUrl(jsConfig.base_url) &&
      supports_join &&
      Boolean(root_keys)

    const asConfig = selectAsConfig()
    const hasAs =
      asConfig.enabled && application_server_address === getHostnameFromUrl(asConfig.base_url)

    const basePath = `/applications/${appId}/devices/${devId}`

    // Prevent default redirect to uplink when tab is already open.
    const payloadFormattersLink = pathname.startsWith(`${basePath}/payload-formatters`)
      ? pathname
      : `${basePath}/payload-formatters`

    const tabs = [
      { title: sharedMessages.overview, name: 'overview', link: basePath },
      { title: sharedMessages.data, name: 'data', link: `${basePath}/data` },
      {
        title: sharedMessages.messages,
        name: 'messages',
        link: `${basePath}/messages`,
        hidden: !mayScheduleDownlinks,
      },
      { title: sharedMessages.location, name: 'location', link: `${basePath}/location` },
      {
        title: sharedMessages.payloadFormatters,
        name: 'develop',
        link: payloadFormattersLink,
        exact: false,
        hidden: !hasAs,
      },
      {
        title: sharedMessages.claiming,
        name: 'claim-auth-code',
        link: `${basePath}/claim-auth-code`,
        hidden: !hasJs,
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
        <EntityTitleSection.Device
          deviceId={devId}
          deviceName={name}
          description={description}
          lastSeen={deviceLastSeen}
          uplinkFrameCount={deviceUplinkFrameCount}
        >
          <Tabs className={style.tabs} narrow tabs={tabs} />
        </EntityTitleSection.Device>
        <Switch>
          <Route exact path={basePath} component={DeviceOverview} />
          <Route exact path={`${basePath}/data`} component={DeviceData} />
          <Route exact path={`${basePath}/messages`} component={DeviceMessages} />
          <Route exact path={`${basePath}/location`} component={DeviceLocation} />
          <Route exact path={`${basePath}/general-settings`} component={DeviceGeneralSettings} />
          {hasAs && (
            <Route path={`${basePath}/payload-formatters`} component={DevicePayloadFormatters} />
          )}
          {hasJs && (
            <Route path={`${basePath}/claim-auth-code`} component={DeviceClaimAuthenticationCode} />
          )}
          <NotFoundRoute />
        </Switch>
      </React.Fragment>
    )
  }
}
