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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import { withSideNavigation } from '../../../components/navigation/side/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { withEnv } from '../../../lib/components/env'
import BreadcrumbView from '../../../lib/components/breadcrumb-view'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import NotFoundRoute from '../../../lib/components/not-found-route'

import OrganizationOverview from '../organization-overview'
import OrganizationData from '../organization-data'
import OrganizationGeneralSettings from '../organization-general-settings'
import OrganizationApiKeys from '../organization-api-keys'
import OrganizationCollaborators from '../organization-collaborators'

@withEnv
@withSideNavigation(function(props) {
  const matchedUrl = props.match.url

  return {
    header: { title: props.orgId, icon: 'organization' },
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
        title: sharedMessages.apiKeys,
        path: `${matchedUrl}/api-keys`,
        icon: 'api_keys',
        exact: false,
      },
      {
        title: sharedMessages.collaborators,
        path: `${matchedUrl}/collaborators`,
        icon: 'organization',
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
@withBreadcrumb('orgs.single', function(props) {
  const { orgId } = props
  return <Breadcrumb path={`/organizations/${orgId}`} icon="organization" content={orgId} />
})
class Organization extends React.Component {
  static propTypes = {
    env: PropTypes.env.isRequired,
    match: PropTypes.match.isRequired,
    orgId: PropTypes.string.isRequired,
    organization: PropTypes.organization.isRequired,
    stopStream: PropTypes.func.isRequired,
  }

  componentWillUnmount() {
    const { orgId, stopStream } = this.props

    stopStream(orgId)
  }

  render() {
    const { match, organization, orgId, env } = this.props

    return (
      <BreadcrumbView>
        <IntlHelmet titleTemplate={`%s - ${organization.name || orgId} - ${env.siteName}`} />
        <Switch>
          <Route exact path={`${match.path}`} component={OrganizationOverview} />
          <Route path={`${match.path}/data`} component={OrganizationData} />
          <Route path={`${match.path}/general-settings`} component={OrganizationGeneralSettings} />
          <Route path={`${match.path}/api-keys`} component={OrganizationApiKeys} />
          <Route path={`${match.path}/collaborators`} component={OrganizationCollaborators} />
          <NotFoundRoute />
        </Switch>
      </BreadcrumbView>
    )
  }
}

export default Organization
