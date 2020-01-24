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

import sharedMessages from '../../../lib/shared-messages'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import SideNavigation from '../../../components/navigation/side'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import IntlHelmet from '../../../lib/components/intl-helmet'
import withRequest from '../../../lib/components/with-request'
import { withEnv } from '../../../lib/components/env'
import Breadcrumbs from '../../../components/breadcrumbs'
import NotFoundRoute from '../../../lib/components/not-found-route'

import ApplicationOverview from '../application-overview'
import ApplicationGeneralSettings from '../application-general-settings'
import ApplicationApiKeys from '../application-api-keys'
import ApplicationLink from '../application-link'
import ApplicationCollaborators from '../application-collaborators'
import ApplicationData from '../application-data'
import ApplicationPayloadFormatters from '../application-payload-formatters'
import ApplicationIntegrationsWebhooks from '../application-integrations-webhooks'
import ApplicationIntegrationsPubsubs from '../application-integrations-pubsubs'
import ApplicationIntegrationsMqtt from '../application-integrations-mqtt'

import {
  getApplication,
  stopApplicationEventsStream,
  getApplicationsRightsList,
} from '../../store/actions/applications'
import {
  selectSelectedApplication,
  selectApplicationFetching,
  selectApplicationError,
  selectApplicationRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '../../store/selectors/applications'
import {
  mayViewApplicationInfo,
  mayViewApplicationEvents,
  mayLinkApplication,
  mayViewApplicationDevices,
  mayCreateOrEditApplicationIntegrations,
  mayEditBasicApplicationInfo,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
} from '../../lib/feature-checks'

import Devices from '../devices'
import PropTypes from '../../../lib/prop-types'

@connect(
  function(state, props) {
    return {
      appId: props.match.params.appId,
      fetching: selectApplicationFetching(state) || selectApplicationRightsFetching(state),
      application: selectSelectedApplication(state),
      error: selectApplicationError(state) || selectApplicationRightsError(state),
      rights: selectApplicationRights(state),
    }
  },
  dispatch => ({
    stopStream: id => dispatch(stopApplicationEventsStream(id)),
    loadData: id => {
      dispatch(getApplication(id, 'name,description'))
      dispatch(getApplicationsRightsList(id))
    },
  }),
)
@withRequest(
  ({ appId, loadData }) => loadData(appId),
  ({ fetching, application }) => fetching || !Boolean(application),
)
@withBreadcrumb('apps.single', function(props) {
  const {
    appId,
    application: { name },
  } = props
  return <Breadcrumb path={`/applications/${appId}`} icon="application" content={name || appId} />
})
@withEnv
export default class Application extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    application: PropTypes.application.isRequired,
    env: PropTypes.env,
    match: PropTypes.match.isRequired,
    stopStream: PropTypes.func.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  static defaultProps = {
    env: undefined,
  }

  componentWillUnmount() {
    const { appId, stopStream } = this.props

    stopStream(appId)
  }

  render() {
    const {
      match: { url: matchedUrl, path },
      application,
      appId,
      env,
      rights,
    } = this.props

    return (
      <React.Fragment>
        <Breadcrumbs />
        <IntlHelmet titleTemplate={`%s - ${application.name || appId} - ${env.siteName}`} />
        <SideNavigation header={{ icon: 'application', title: application.name || appId }}>
          {mayViewApplicationInfo.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.overview}
              path={matchedUrl}
              icon="overview"
              exact
            />
          )}
          {mayViewApplicationDevices.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.devices}
              path={`${matchedUrl}/devices`}
              icon="devices"
            />
          )}
          {mayViewApplicationEvents.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.data}
              path={`${matchedUrl}/data`}
              icon="data"
            />
          )}
          {mayLinkApplication.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.link}
              path={`${matchedUrl}/link`}
              icon="link"
            />
          )}
          {mayLinkApplication.check(rights) && (
            <SideNavigation.Item title={sharedMessages.payloadFormatters} icon="code">
              <SideNavigation.Item
                title={sharedMessages.uplink}
                path={`${matchedUrl}/payload-formatters/uplink`}
                icon="uplink"
              />
              <SideNavigation.Item
                title={sharedMessages.downlink}
                path={`${matchedUrl}/payload-formatters/downlink`}
                icon="downlink"
              />
            </SideNavigation.Item>
          )}
          {mayCreateOrEditApplicationIntegrations.check(rights) && (
            <SideNavigation.Item title={sharedMessages.integrations} icon="integration">
              <SideNavigation.Item
                title={sharedMessages.mqtt}
                path={`${matchedUrl}/integrations/mqtt`}
                icon="extension"
              />
              <SideNavigation.Item
                title={sharedMessages.webhooks}
                path={`${matchedUrl}/integrations/webhooks`}
                icon="extension"
              />
              <SideNavigation.Item
                title={sharedMessages.pubsubs}
                path={`${matchedUrl}/integrations/pubsubs`}
                icon="extension"
              />
            </SideNavigation.Item>
          )}
          {mayViewOrEditApplicationCollaborators.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.collaborators}
              path={`${matchedUrl}/collaborators`}
              icon="organization"
            />
          )}
          {mayViewOrEditApplicationApiKeys.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.apiKeys}
              path={`${matchedUrl}/api-keys`}
              icon="api_keys"
            />
          )}
          {mayEditBasicApplicationInfo.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.generalSettings}
              path={`${matchedUrl}/general-settings`}
              icon="general_settings"
            />
          )}
        </SideNavigation>
        <Switch>
          <Route exact path={`${path}`} component={ApplicationOverview} />
          <Route path={`${path}/general-settings`} component={ApplicationGeneralSettings} />
          <Route path={`${path}/api-keys`} component={ApplicationApiKeys} />
          <Route path={`${path}/link`} component={ApplicationLink} />
          <Route path={`${path}/devices`} component={Devices} />
          <Route path={`${path}/collaborators`} component={ApplicationCollaborators} />
          <Route path={`${path}/data`} component={ApplicationData} />
          <Route path={`${path}/payload-formatters`} component={ApplicationPayloadFormatters} />
          <Route path={`${path}/integrations/mqtt`} component={ApplicationIntegrationsMqtt} />
          <Route
            path={`${path}/integrations/webhooks`}
            component={ApplicationIntegrationsWebhooks}
          />
          <Route path={`${path}/integrations/pubsubs`} component={ApplicationIntegrationsPubsubs} />
          <NotFoundRoute />
        </Switch>
      </React.Fragment>
    )
  }
}
