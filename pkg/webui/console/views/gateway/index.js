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
import { replace } from 'connected-react-router'

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import { withSideNavigation } from '../../../components/navigation/side/context'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'

import GatewayOverview from '../gateway-overview'
import GatewayApiKeys from '../gateway-api-keys'
import GatewayCollaborators from '../gateway-collaborators'
import GatewayLocation from '../gateway-location'
import GatewayData from '../gateway-data'
import GatewayGeneralSettings from '../gateway-general-settings'

import {
  getGateway,
  startGatewayEventsStream,
  stopGatewayEventsStream,
} from '../../store/actions/gateways'
import {
  selectGatewayFetching,
  selectGatewayError,
  selectSelectedGateway,
} from '../../store/selectors/gateways'
import withEnv, { EnvProvider } from '../../../lib/components/env'

@connect(function (state, props) {
  const gtwId = props.match.params.gtwId

  return {
    gtwId,
    gateway: selectSelectedGateway(state),
    error: selectGatewayError(state),
    fetching: selectGatewayFetching(state),
  }
},
dispatch => ({
  getGateway: (id, meta) => dispatch(getGateway(id, meta)),
  startStream: id => dispatch(startGatewayEventsStream(id)),
  stopStream: id => dispatch(stopGatewayEventsStream(id)),
  redirectToList: () => dispatch(replace('/console/gateways')),
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
        title: sharedMessages.data,
        path: `${matchedUrl}/data`,
        icon: 'data',
      },
      {
        title: sharedMessages.location,
        path: `${matchedUrl}/location`,
        icon: 'location',
      },
      {
        title: sharedMessages.collaborators,
        path: `${matchedUrl}/collaborators`,
        icon: 'organization',
        exact: false,
      },
      {
        title: sharedMessages.apiKeys,
        path: `${matchedUrl}/api-keys`,
        icon: 'api_keys',
        exact: false,
      },
      {
        title: sharedMessages.generalSettings,
        path: `${matchedUrl}/general-settings`,
        icon: 'general_settings',
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
@withEnv
export default class Gateway extends React.Component {

  componentDidMount () {
    const { getGateway, startStream, match } = this.props
    const { gtwId } = match.params

    startStream(gtwId)
    getGateway(gtwId, [
      'name',
      'description',
      'enforce_duty_cycle',
      'frequency_plan_id',
      'gateway_server_address',
      'enforce_duty_cycle',
      'antennas',
    ])
  }

  componentDidUpdate (prevProps) {
    const { gateway, redirectToList } = this.props

    if (Boolean(prevProps.gateway) && !Boolean(gateway)) {
      redirectToList()
    }
  }

  componentWillUnmount () {
    const { stopStream, gtwId } = this.props

    stopStream(gtwId)
  }

  render () {
    const { fetching, error, match, gateway, gtwId, env } = this.props

    // show any gateway fetching error, e.g. not found, no rights, etc
    if (error) {
      throw error
    }

    if (fetching || !gateway) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    return (
      <EnvProvider env={env}>
        <IntlHelmet
          titleTemplate={`%s - ${gateway.name || gtwId} - ${env.site_name}`}
        />
        <Switch>
          <Route exact path={`${match.path}`} component={GatewayOverview} />
          <Route path={`${match.path}/api-keys`} component={GatewayApiKeys} />
          <Route path={`${match.path}/collaborators`} component={GatewayCollaborators} />
          <Route path={`${match.path}/location`} component={GatewayLocation} />
          <Route path={`${match.path}/data`} component={GatewayData} />
          <Route path={`${match.path}/general-settings`} component={GatewayGeneralSettings} />
        </Switch>
      </EnvProvider>
    )
  }
}
