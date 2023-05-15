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

import React, { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'

import organizationIcon from '@assets/misc/organization.svg'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'
import SideNavigation from '@ttn-lw/components/navigation/side'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import OrganizationOverview from '@console/views/organization-overview'
import OrganizationData from '@console/views/organization-data'
import OrganizationGeneralSettings from '@console/views/organization-general-settings'
import OrganizationApiKeys from '@console/views/organization-api-keys'
import OrganizationCollaborators from '@console/views/organization-collaborators'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const Organization = props => {
  const {
    match: { url: matchedUrl, path },
    siteName,
    organization,
    orgId,
    stopStream,
    mayViewInformation,
    mayEditInformation,
    mayViewOrEditApiKeys,
    mayViewOrEditCollaborators,
  } = props
  const { name: organizationName } = organization

  useBreadcrumbs(
    'orgs.single',
    <Breadcrumb path={`/organizations/${orgId}`} content={organizationName || orgId} />,
  )

  useEffect(
    () => () => {
      stopStream(orgId)
    },
    [orgId, stopStream],
  )

  return (
    <React.Fragment>
      <Breadcrumbs />
      <IntlHelmet titleTemplate={`%s - ${organization.name || orgId} - ${siteName}`} />
      <SideNavigation
        header={{
          title: organization.name || orgId,
          icon: organizationIcon,
          iconAlt: sharedMessages.organization,
          to: matchedUrl,
        }}
      >
        {mayViewInformation && (
          <SideNavigation.Item
            title={sharedMessages.overview}
            icon="overview"
            path={matchedUrl}
            exact
          />
        )}
        <SideNavigation.Item
          title={sharedMessages.liveData}
          icon="data"
          path={`${matchedUrl}/data`}
        />
        {mayViewOrEditApiKeys && (
          <SideNavigation.Item
            title={sharedMessages.apiKeys}
            icon="api_keys"
            path={`${matchedUrl}/api-keys`}
          />
        )}
        {mayViewOrEditCollaborators && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            icon="collaborators"
            path={`${matchedUrl}/collaborators`}
          />
        )}
        {mayEditInformation && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            icon="general_settings"
            path={`${matchedUrl}/general-settings`}
          />
        )}
      </SideNavigation>
      <Routes>
        <Route exact path={`${path}`} component={OrganizationOverview} />
        <Route path={`${path}/data`} component={OrganizationData} />
        <Route path={`${path}/general-settings`} component={OrganizationGeneralSettings} />
        <Route path={`${path}/api-keys`} component={OrganizationApiKeys} />
        <Route path={`${path}/collaborators`} component={OrganizationCollaborators} />
        <NotFoundRoute />
      </Routes>
    </React.Fragment>
  )
}

Organization.propTypes = {
  match: PropTypes.match.isRequired,
  mayEditInformation: PropTypes.bool.isRequired,
  mayViewInformation: PropTypes.bool.isRequired,
  mayViewOrEditApiKeys: PropTypes.bool.isRequired,
  mayViewOrEditCollaborators: PropTypes.bool.isRequired,
  orgId: PropTypes.string.isRequired,
  organization: PropTypes.organization.isRequired,
  siteName: PropTypes.string.isRequired,
  stopStream: PropTypes.func.isRequired,
}

export default Organization
