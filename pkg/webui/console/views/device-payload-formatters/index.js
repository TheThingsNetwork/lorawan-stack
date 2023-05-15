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
import { Routes, Route, Navigate } from 'react-router-dom'

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Tab from '@ttn-lw/components/tabs'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import DeviceUplinkPayloadFormatters from '@console/containers/device-payload-formatters/uplink'
import DeviceDownlinkPayloadFormatters from '@console/containers/device-payload-formatters/downlink'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectSelectedDeviceId } from '@console/store/selectors/devices'

import style from './device-payload-formatters.styl'

@connect(state => ({
  appId: selectSelectedApplicationId(state),
  devId: selectSelectedDeviceId(state),
}))
@withBreadcrumb('device.single.payload-formatters', props => {
  const { appId, devId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/payload-formatters`}
      content={sharedMessages.payloadFormatters}
    />
  )
})
export default class DevicePayloadFormatters extends Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }

  render() {
    const {
      match: { url },
    } = this.props

    const tabs = [
      { title: sharedMessages.uplink, name: 'uplink', link: `${url}/uplink` },
      { title: sharedMessages.downlink, name: 'downlink', link: `${url}/downlink` },
    ]

    return (
      <div className={style.fullWidth}>
        <Tab className={style.tabs} tabs={tabs} divider />
        <Routes>
          <Route path="*" element={<Navigate to={`${url}uplink`} replace />} />
          <Route exact path={`${url}/uplink`} component={DeviceUplinkPayloadFormatters} />
          <Route exact path={`${url}/downlink`} component={DeviceDownlinkPayloadFormatters} />
          <NotFoundRoute />
        </Routes>
      </div>
    )
  }
}
