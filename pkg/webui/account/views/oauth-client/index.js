// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import applicationIcon from '@assets/misc/application.svg'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import withRequest from '@ttn-lw/lib/components/with-request'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import SideNavigation from '@ttn-lw/components/navigation/side'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import OAuthClientOverview from '@account/views/oauth-client-overview'
/* import ApplicationGeneralSettings from '@console/views/application-general-settings'
import ApplicationCollaborators from '@console/views/application-collaborators' */

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayPerformAdminActions } from '@account/lib/feature-checks'


const OAuthClient = props => {
  const {
    appId,
    match: { url: matchedUrl, path },
    application,
    siteName,
  } = props

  const name = application.name || appId

  useBreadcrumbs('apps.single', <Breadcrumb path={`/applications/${appId}`} content={name} />)

  return (
    <React.Fragment>
      <Breadcrumbs />
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <SideNavigation
        header={{
          icon: applicationIcon,
          iconAlt: sharedMessages.application,
          title: name,
          to: matchedUrl,
        }}
      >
        {mayPerformAdminActions && (
          <SideNavigation.Item
            title={sharedMessages.overview}
            path={matchedUrl}
            icon="overview"
            exact
          />
        )}
        {mayPerformAdminActions && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path={`${matchedUrl}/collaborators`}
            icon="organization"
          />
        )}
        {mayPerformAdminActions && (
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
        <Route path={`${path}/collaborators`} component={ApplicationCollaborators} />
        <NotFoundRoute />
      </Switch>
    </React.Fragment>
  )
}

OAuthClient.propTypes = {
  appId: PropTypes.string.isRequired,
  application: PropTypes.application.isRequired,
  match: PropTypes.match.isRequired,
  siteName: PropTypes.string.isRequired,
}

export default connect(
  state => ({
    oauthClientId: props.match.params.appId,
    fetching: selectApplicationFetching(state) || selectApplicationRightsFetching(state),
    oauthClient: selectSelectedApplication(state),
    error: selectApplicationError(state) || selectApplicationRightsError(state),
    siteName: selectApplicationSiteName(),
  }),
  dispatch => ({
    loadData: id => {
      dispatch(getOAuthClient(id, 'name,description,secret,state,state_description'))
    },
  }),
)(
  withRequest(
    ({ oauthClientId, loadData }) => loadData(oauthClientId),
    ({ fetching, oauthClient }) => fetching || !Boolean(oauthClient),
  )(OAuthClient),
)
