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

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'
import SideNavigation from '@ttn-lw/components/navigation/side'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import withRequest from '@ttn-lw/lib/components/with-request'
import withEnv from '@ttn-lw/lib/components/env'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import GatewayCollaborators from '@console/views/gateway-collaborators'
import GatewayLocation from '@console/views/gateway-location'
import GatewayData from '@console/views/gateway-data'
import GatewayGeneralSettings from '@console/views/gateway-general-settings'
import GatewayApiKeys from '@console/views/gateway-api-keys'
import GatewayOverview from '@console/views/gateway-overview'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayViewGatewayInfo,
  mayViewGatewayEvents,
  mayViewOrEditGatewayLocation,
  mayViewOrEditGatewayCollaborators,
  mayViewOrEditGatewayApiKeys,
  mayEditBasicGatewayInformation,
} from '@console/lib/feature-checks'

import {
  getGateway,
  stopGatewayEventsStream,
  getGatewaysRightsList,
} from '@console/store/actions/gateways'

import {
  selectGatewayFetching,
  selectGatewayError,
  selectSelectedGateway,
  selectGatewayRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
} from '@console/store/selectors/gateways'

@connect(
  function (state, props) {
    const gtwId = props.match.params.gtwId
    const gateway = selectSelectedGateway(state)

    return {
      gtwId,
      gateway,
      error: selectGatewayError(state) || selectGatewayRightsError(state),
      fetching: selectGatewayFetching(state) || selectGatewayRightsFetching(state),
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
      'antennas',
      'location_public',
      'status_public',
      'auto_update',
      'schedule_downlink_late',
      'update_location_from_status',
      'update_channel',
      'schedule_anytime_delay',
      'attributes',
    ]),
  ({ fetching, gateway }) => fetching || !Boolean(gateway),
)
@withBreadcrumb('gateways.single', function (props) {
  const {
    gtwId,
    gateway: { name },
  } = props

  return <Breadcrumb path={`/gateways/${gtwId}`} content={name || gtwId} />
})
@withEnv
export default class Gateway extends React.Component {
  static propTypes = {
    env: PropTypes.env,
    gateway: PropTypes.gateway.isRequired,
    gtwId: PropTypes.string.isRequired,
    match: PropTypes.match.isRequired,
    rights: PropTypes.rights.isRequired,
    stopStream: PropTypes.func.isRequired,
  }

  static defaultProps = {
    env: undefined,
  }

  componentWillUnmount() {
    const { stopStream, gtwId } = this.props

    stopStream(gtwId)
  }

  render() {
    const {
      match: { url: matchedUrl },
      gateway,
      gtwId,
      env,
      rights,
      match,
    } = this.props

    return (
      <React.Fragment>
        <Breadcrumbs />
        <IntlHelmet titleTemplate={`%s - ${gateway.name || gtwId} - ${env.siteName}`} />
        <SideNavigation header={{ icon: 'gateway', title: gateway.name || gtwId, to: matchedUrl }}>
          {mayViewGatewayInfo.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.overview}
              path={matchedUrl}
              icon="overview"
              exact
            />
          )}
          {mayViewGatewayEvents.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.liveData}
              path={`${matchedUrl}/data`}
              icon="data"
            />
          )}
          {mayViewOrEditGatewayLocation.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.location}
              path={`${matchedUrl}/location`}
              icon="location"
            />
          )}
          {mayViewOrEditGatewayCollaborators.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.collaborators}
              path={`${matchedUrl}/collaborators`}
              icon="organization"
            />
          )}
          {mayViewOrEditGatewayApiKeys.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.apiKeys}
              path={`${matchedUrl}/api-keys`}
              icon="api_keys"
            />
          )}
          {mayEditBasicGatewayInformation.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.generalSettings}
              path={`${matchedUrl}/general-settings`}
              icon="general_settings"
            />
          )}
        </SideNavigation>
        <Switch>
          <Route exact path={`${match.path}`} component={GatewayOverview} />
          <Route path={`${match.path}/api-keys`} component={GatewayApiKeys} />
          <Route path={`${match.path}/collaborators`} component={GatewayCollaborators} />
          <Route path={`${match.path}/location`} component={GatewayLocation} />
          <Route path={`${match.path}/data`} component={GatewayData} />
          <Route path={`${match.path}/general-settings`} component={GatewayGeneralSettings} />
          <NotFoundRoute />
        </Switch>
      </React.Fragment>
    )
  }
}
