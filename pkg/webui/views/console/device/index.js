// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import Spinner from '../../../components/spinner'

import DeviceOverview from '../device-overview'

import { getDevice } from '../../../actions/device'

@connect(function ({ device }, props) {
  return {
    devId: props.match.params.devId,
    fetching: device.fetching,
    error: device.error,
  }
})
@withBreadcrumb('device.single', function (props) {
  const { appId, devId } = props
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
    const { dispatch, devId } = this.props

    dispatch(getDevice(devId))
  }

  render () {
    const { fetching, error, match } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    // show any device fetching error, e.g. not found, no rights, etc
    if (error) {
      return 'ERROR'
    }

    return (
      <Switch>
        <Route exact path={`${match.path}`} component={DeviceOverview} />
      </Switch>
    )
  }
}
