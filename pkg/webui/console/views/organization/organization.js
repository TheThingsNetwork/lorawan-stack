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

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'
import SideNavigation from '@ttn-lw/components/navigation/side'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import { withEnv } from '@ttn-lw/lib/components/env'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import OrganizationOverview from '@console/views/organization-overview'
import OrganizationData from '@console/views/organization-data'
import OrganizationGeneralSettings from '@console/views/organization-general-settings'
import OrganizationApiKeys from '@console/views/organization-api-keys'
import OrganizationCollaborators from '@console/views/organization-collaborators'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayViewOrganizationInformation,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
  mayEditBasicOrganizationInformation,
} from '@console/lib/feature-checks'

@withEnv
@withBreadcrumb('orgs.single', function(props) {
  const {
    orgId,
    organization: { name },
  } = props
  return <Breadcrumb path={`/organizations/${orgId}`} content={name || orgId} />
})
class Organization extends React.Component {
  static propTypes = {
    env: PropTypes.env.isRequired,
    match: PropTypes.match.isRequired,
    orgId: PropTypes.string.isRequired,
    organization: PropTypes.organization.isRequired,
    rights: PropTypes.rights.isRequired,
    stopStream: PropTypes.func.isRequired,
  }

  componentWillUnmount() {
    const { orgId, stopStream } = this.props

    stopStream(orgId)
  }

  render() {
    const {
      match: { url: matchedUrl, path },
      organization,
      orgId,
      env,
      rights,
    } = this.props

    return (
      <React.Fragment>
        <Breadcrumbs />
        <IntlHelmet titleTemplate={`%s - ${organization.name || orgId} - ${env.siteName}`} />
        <SideNavigation header={{ title: organization.name || orgId, icon: 'organization' }}>
          {mayViewOrganizationInformation.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.overview}
              icon="overview"
              path={matchedUrl}
              exact
            />
          )}
          <SideNavigation.Item
            title={sharedMessages.data}
            icon="data"
            path={`${matchedUrl}/data`}
          />
          {mayViewOrEditOrganizationApiKeys.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.apiKeys}
              icon="api_keys"
              path={`${matchedUrl}/api-keys`}
            />
          )}
          {mayViewOrEditOrganizationCollaborators.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.collaborators}
              icon="collaborators"
              path={`${matchedUrl}/collaborators`}
            />
          )}
          {mayEditBasicOrganizationInformation.check(rights) && (
            <SideNavigation.Item
              title={sharedMessages.generalSettings}
              icon="general_settings"
              path={`${matchedUrl}/general-settings`}
            />
          )}
        </SideNavigation>
        <Switch>
          <Route exact path={`${path}`} component={OrganizationOverview} />
          <Route path={`${path}/data`} component={OrganizationData} />
          <Route path={`${path}/general-settings`} component={OrganizationGeneralSettings} />
          <Route path={`${path}/api-keys`} component={OrganizationApiKeys} />
          <Route path={`${path}/collaborators`} component={OrganizationCollaborators} />
          <NotFoundRoute />
        </Switch>
      </React.Fragment>
    )
  }
}

export default Organization
