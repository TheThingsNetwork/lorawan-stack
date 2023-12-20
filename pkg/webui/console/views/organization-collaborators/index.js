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
import { Routes, Route, useParams } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import SubViewError from '@console/views/sub-view-error'
import OrganizationCollaboratorsList from '@console/views/organization-collaborators-list'
import OrganizationCollaboratorAdd from '@console/views/organization-collaborator-add'
import OrganizationCollaboratorEdit from '@console/views/organization-collaborator-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userPathId as userPathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayViewOrEditOrganizationCollaborators } from '@console/lib/feature-checks'

const OrganizationCollaborators = () => {
  const { orgId } = useParams()

  useBreadcrumbs(
    'orgs.single.collaborators',
    <Breadcrumb
      path={`/organizations/${orgId}/collaborators`}
      content={sharedMessages.collaborators}
    />,
  )

  return (
    <Require
      featureCheck={mayViewOrEditOrganizationCollaborators}
      otherwise={{ redirect: `/organizations/${orgId}` }}
    >
      <ErrorView errorRender={SubViewError}>
        <Routes>
          <Route index Component={OrganizationCollaboratorsList} />
          <Route path="add" Component={OrganizationCollaboratorAdd} />
          <Route
            path="user/:collaboratorId"
            element={
              <ValidateRouteParam
                check={{ collaboratorId: userPathIdRegexp }}
                Component={OrganizationCollaboratorEdit}
              />
            }
          />
          <Route path="*" element={<GenericNotFound />} />
        </Routes>
      </ErrorView>
    </Require>
  )
}

export default OrganizationCollaborators
