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
import { Switch, Route } from 'react-router'
import { connect } from 'react-redux'

import sharedMessages from '../../../lib/shared-messages'
import { withSideNavigation } from '../../../components/navigation/side/context'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'

import GatewayOverview from '../gateway-overview'
import GatewayApiKeys from '../gateway-api-keys'

import {
  getGateway,
  startGatewayEventsStream,
  stopGatewayEventsStream,
} from '../../store/actions/gateway'
import {
  fetchingSelector,
  errorSelector,
  gatewaySelector,
} from '../../store/selectors/gateway'

@connect(function (state, props) {
  const gtwId = props.match.params.gtwId

  return {
    gtwId,
    gateway: gatewaySelector(state, props),
    error: errorSelector(state, props),
    fetching: fetchingSelector(state, props),
  }
},
dispatch => ({
  getGateway: (id, meta) => dispatch(getGateway(id, meta)),
  startStream: id => dispatch(startGatewayEventsStream(id)),
  stopStream: id => dispatch(stopGatewayEventsStream(id)),
}))
@withSideNavigation(function (props) {
  const { match, gtwId } = props
  const matchedUrl = match.url

  return {
    header: { title: gtwId, icon: 'gateway' },
    entries: [
      {
        title: sharedMessages.overview,
        path: matchedUrl,
        icon: 'overview',
      },
      {
        title: sharedMessages.apiKeys,
        path: `${matchedUrl}/api-keys`,
        icon: 'api_keys',
        exact: false,
      },
    ],
  }
})
@withBreadcrumb('gateways.single', function (props) {
  const { gtwId } = props

  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}`}
      icon="gateway"
      content={gtwId}
    />
  )
})
export default class Gateway extends React.Component {

  componentDidMount () {
    const { getGateway, startStream, match } = this.props
    const { gtwId } = match.params

    startStream(gtwId)
    getGateway(gtwId, {
      selectors: [
        'name',
        'description',
        'enforce_duty_cycle',
        'frequency_plan_id',
        'gateway_server_address',
        'enforce_duty_cycle',
      ],
    })
  }

  componentWillUnmount () {
    const { stopStream, gtwId } = this.props

    stopStream(gtwId)
  }

  render () {
    const { fetching, error, match, gateway } = this.props

    // show any gateway fetching error, e.g. not found, no rights, etc
    if (error) {
      return 'ERROR'
    }

    if (fetching || !gateway) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    return (
      <Switch>
        <Route exact path={`${match.path}`} component={GatewayOverview} />
        <Route path={`${match.path}/api-keys`} component={GatewayApiKeys} />
      </Switch>
    )
  }
}
