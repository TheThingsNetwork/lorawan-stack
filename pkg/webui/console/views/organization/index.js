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

import React, { useEffect } from 'react'
import { Routes, Route, useParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import Require from '@console/lib/components/require'

import OrganizationOverview from '@console/views/organization-overview'
import OrganizationData from '@console/views/organization-data'
import OrganizationGeneralSettings from '@console/views/organization-general-settings'
import OrganizationApiKeys from '@console/views/organization-api-keys'
import OrganizationCollaborators from '@console/views/organization-collaborators'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import { mayViewOrganizationsOfUser } from '@console/lib/feature-checks'

import {
  getOrganization,
  stopOrganizationEventsStream,
  getOrganizationsRightsList,
} from '@console/store/actions/organizations'

import { selectSelectedOrganization } from '@console/store/selectors/organizations'

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
  const dispatch = useDispatch()
  const siteName = selectApplicationSiteName()

  useBreadcrumbs('orgs.single', <Breadcrumb path={`/organizations/${orgId}`} content={name} />)

  useEffect(
    () => () => {
      dispatch(stopOrganizationEventsStream(orgId))
    },
    [dispatch, orgId],
  )

  return (
    <React.Fragment>
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <Routes>
        <Route index Component={OrganizationOverview} />
        <Route path="data" Component={OrganizationData} />
        <Route path="general-settings" Component={OrganizationGeneralSettings} />
        <Route path="api-keys/*" Component={OrganizationApiKeys} />
        <Route path="collaborators/*" Component={OrganizationCollaborators} />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </React.Fragment>
  )
}

export default Organization
