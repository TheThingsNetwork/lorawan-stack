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

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import { withSideNavigation } from '../../../components/navigation/side/context'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import withRequest from '../../../lib/components/with-request'
import withEnv from '../../../lib/components/env'
import BreadcrumbView from '../../../lib/components/breadcrumb-view'
import NotFoundRoute from '../../../lib/components/not-found-route'

import GatewayOverview from '../gateway-overview'
import GatewayApiKeys from '../gateway-api-keys'
import GatewayCollaborators from '../gateway-collaborators'
import GatewayLocation from '../gateway-location'
import GatewayData from '../gateway-data'
import GatewayGeneralSettings from '../gateway-general-settings'

import {
  getGateway,
  stopGatewayEventsStream,
  getGatewaysRightsList,
} from '../../store/actions/gateways'
import {
  selectGatewayFetching,
  selectGatewayError,
  selectSelectedGateway,
  selectGatewayRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
} from '../../store/selectors/gateways'
import {
  mayViewGatewayInfo,
  mayViewGatewayEvents,
  mayViewOrEditGatewayLocation,
  mayViewOrEditGatewayCollaborators,
  mayViewOrEditGatewayApiKeys,
  mayEditBasicGatewayInformation,
} from '../../lib/feature-checks'

@connect(
  function(state, props) {
    const gtwId = props.match.params.gtwId
    const gateway = selectSelectedGateway(state)

    return {
      gtwId,
      gateway,
      error: selectGatewayError(state) && selectGatewayRightsError(state),
      fetching: selectGatewayFetching(state) && selectGatewayRightsFetching(state),
      rights: selectGatewayRights(state),
    }
  },
  dispatch => ({
    loadData: (id, meta) => {
      dispatch(getGateway(id, meta))
      dispatch(getGatewaysRightsList(id))
    },
    stopStream: id => dispatch(stopGatewayEventsStream(id)),
  }),
)
@withRequest(
  ({ gtwId, loadData }) =>
    loadData(gtwId, [
      'name',
      'description',
      'enforce_duty_cycle',
      'frequency_plan_id',
      'gateway_server_address',
      'enforce_duty_cycle',
      'antennas',
    ]),
  ({ fetching, gateway }) => fetching || !Boolean(gateway),
)
@withSideNavigation(function(props) {
  const {
    match: { url: matchedUrl },
    gtwId,
    rights,
  } = props

  return {
    header: { title: gtwId, icon: 'gateway' },
    entries: [
      {
        title: sharedMessages.overview,
        path: matchedUrl,
        icon: 'overview',
        hidden: !mayViewGatewayInfo.check(rights),
      },
      {
        title: sharedMessages.data,
        path: `${matchedUrl}/data`,
        icon: 'data',
        hidden: !mayViewGatewayEvents.check(rights),
      },
      {
        title: sharedMessages.location,
        path: `${matchedUrl}/location`,
        icon: 'location',
        hidden: !mayViewOrEditGatewayLocation.check(rights),
      },
      {
        title: sharedMessages.collaborators,
        path: `${matchedUrl}/collaborators`,
        icon: 'organization',
        exact: false,
        hidden: !mayViewOrEditGatewayCollaborators.check(rights),
      },
      {
        title: sharedMessages.apiKeys,
        path: `${matchedUrl}/api-keys`,
        icon: 'api_keys',
        exact: false,
        hidden: !mayViewOrEditGatewayApiKeys.check(rights),
      },
      {
        title: sharedMessages.generalSettings,
        path: `${matchedUrl}/general-settings`,
        icon: 'general_settings',
        hidden: !mayEditBasicGatewayInformation.check(rights),
      },
    ],
  }
})
@withBreadcrumb('gateways.single', function(props) {
  const { gtwId } = props

  return <Breadcrumb path={`/gateways/${gtwId}`} icon="gateway" content={gtwId} />
})
@withEnv
export default class Gateway extends React.Component {
  render() {
    const { match, gateway, gtwId, env } = this.props

    return (
      <BreadcrumbView>
        <IntlHelmet titleTemplate={`%s - ${gateway.name || gtwId} - ${env.siteName}`} />
        <Switch>
          <Route exact path={`${match.path}`} component={GatewayOverview} />
          <Route path={`${match.path}/api-keys`} component={GatewayApiKeys} />
          <Route path={`${match.path}/collaborators`} component={GatewayCollaborators} />
          <Route path={`${match.path}/location`} component={GatewayLocation} />
          <Route path={`${match.path}/data`} component={GatewayData} />
          <Route path={`${match.path}/general-settings`} component={GatewayGeneralSettings} />
          <NotFoundRoute />
        </Switch>
      </BreadcrumbView>
    )
  }
}
