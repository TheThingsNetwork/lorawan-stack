// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route, useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Tabs from '@ttn-lw/components/tabs'
import { IconCollaborators, IconGeneralSettings, IconKey } from '@ttn-lw/components/icon'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import Require from '@console/lib/components/require'

import OrganizationGeneralSettings from '@console/views/organization-general-settings'
import OrganizationApiKeys from '@console/views/organization-api-keys'
import OrganizationCollaborators from '@console/views/organization-collaborators'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewOrganizationsOfUser } from '@console/lib/feature-checks'

import { getOrganization, getOrganizationsRightsList } from '@console/store/actions/organizations'

import { selectSelectedOrganization } from '@console/store/selectors/organizations'

import OrganizationHeader from './organization-header'

const Organization = () => {
  const { orgId } = useParams()

  // Check whether application still exists after it has been possibly deleted.
  const organization = useSelector(selectSelectedOrganization)
  const hasOrganization = Boolean(organization)

  return (
    <Require featureCheck={mayViewOrganizationsOfUser} otherwise={{ redirect: '/' }}>
      <RequireRequest
        requestAction={[
          getOrganization(
            orgId,
            'name,description,administrative_contact,technical_contact,fanout_notifications',
          ),
          getOrganizationsRightsList(orgId),
        ]}
      >
        {hasOrganization && <OrganizationInner />}
      </RequireRequest>
    </Require>
  )
}

const OrganizationInner = () => {
  const { orgId } = useParams()
  const organization = useSelector(selectSelectedOrganization)
  const name = organization.name || orgId
  const siteName = selectApplicationSiteName()

  useBreadcrumbs(
    'overview.orgs.single',
    <Breadcrumb path={`/organizations/${orgId}`} content={name} />,
  )

  const basePath = `/organizations/${orgId}`

  const tabs = [
    {
      title: sharedMessages.members,
      name: 'members',
      link: basePath,
      icon: IconCollaborators,
    },
    {
      title: sharedMessages.apiKeys,
      name: 'api-keys',
      link: `${basePath}/api-keys`,
      icon: IconKey,
    },
    {
      title: sharedMessages.settings,
      name: 'general-settings',
      link: `${basePath}/general-settings`,
      icon: IconGeneralSettings,
    },
  ]

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <OrganizationHeader org={organization} />
      <Tabs tabs={tabs} divider />
      <Routes>
        <Route path="/*" Component={OrganizationCollaborators} />
        <Route path="general-settings" Component={OrganizationGeneralSettings} />
        <Route path="api-keys/*" Component={OrganizationApiKeys} />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </>
  )
}

export default Organization
