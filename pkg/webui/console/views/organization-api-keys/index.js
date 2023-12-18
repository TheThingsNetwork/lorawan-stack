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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import SubViewError from '@console/views/sub-view-error'
import OrganizationApiKeysList from '@console/views/organization-api-keys-list'
import OrganizationApiKeyAdd from '@console/views/organization-api-key-add'
import OrganizationApiKeyEdit from '@console/views/organization-api-key-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { apiKeyPath as apiKeyPathRegexp } from '@console/lib/regexp'
import { mayViewOrEditOrganizationApiKeys } from '@console/lib/feature-checks'

const OrganizationApiKeys = () => {
  const { orgId } = useParams()

  useBreadcrumbs('org.single.api-keys', [
    {
      path: `/organizations/${orgId}/api-keys`,
      content: sharedMessages.apiKeys,
    },
  ])

  return (
    <Require
      featureCheck={mayViewOrEditOrganizationApiKeys}
      otherwise={{ redirect: `/organizations/${orgId}` }}
    >
      <ErrorView errorRender={SubViewError}>
        <Routes>
          <Route index Component={OrganizationApiKeysList} />
          <Route path="add" Component={OrganizationApiKeyAdd} />
          <Route
            path=":apiKeyId"
            element={
              <ValidateRouteParam
                check={{ apiKeyId: apiKeyPathRegexp }}
                Component={OrganizationApiKeyEdit}
              />
            }
          />
          <Route path="*" element={<GenericNotFound />} />
        </Routes>
      </ErrorView>
    </Require>
  )
}

export default OrganizationApiKeys
