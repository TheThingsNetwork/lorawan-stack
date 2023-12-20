// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'

import Require from '@console/lib/components/require'

import Organization from '@console/views/organization'
import OrganizationAdd from '@console/views/organization-add'
import OrganizationsList from '@console/views/organizations-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayViewOrganizationsOfUser } from '@console/lib/feature-checks'

const Organizations = () => {
  useBreadcrumbs(
    'orgs',
    <Breadcrumb path="/organizations" content={sharedMessages.organizations} />,
  )

  return (
    <Require featureCheck={mayViewOrganizationsOfUser} otherwise={{ redirect: '/' }}>
      <Routes>
        <Route index Component={OrganizationsList} />
        <Route path="add" Component={OrganizationAdd} />
        <Route
          path=":orgId/*"
          element={<ValidateRouteParam check={{ orgId: pathIdRegexp }} Component={Organization} />}
        />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </Require>
  )
}

export default Organizations
