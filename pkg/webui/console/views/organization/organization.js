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

import OrganizationOverview from '../organization-overview'

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
        </Switch>
      </BreadcrumbView>
    )
  }
}

export default Organization
