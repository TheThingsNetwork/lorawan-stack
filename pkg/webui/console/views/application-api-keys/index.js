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
import { Routes, Route, useParams } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import ApplicationApiKeysList from '@console/views/application-api-keys-list'
import ApplicationApiKeyAdd from '@console/views/application-api-key-add'
import SubViewError from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { apiKeyPath as apiKeyPathRegexp } from '@console/lib/regexp'
import { mayViewOrEditApplicationApiKeys } from '@console/lib/feature-checks'

import ApplicationApiKeyEdit from '../application-api-key-edit'

const ApplicationApiKeys = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.api-keys',
    <Breadcrumb path={`/applications/${appId}/api-keys`} content={sharedMessages.apiKeys} />,
  )

  return (
    <Require
      featureCheck={mayViewOrEditApplicationApiKeys}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      <ErrorView errorRender={SubViewError}>
        <Routes>
          <Route index Component={ApplicationApiKeysList} />
          <Route path="/add" Component={ApplicationApiKeyAdd} />
          <Route
            path="/:apiKeyId"
            element={
              <ValidateRouteParam
                check={{ apiKeyId: apiKeyPathRegexp }}
                Component={ApplicationApiKeyEdit}
              />
            }
          />
          <Route path="*" element={<GenericNotFound />} />
        </Routes>
      </ErrorView>
    </Require>
  )
}

export default ApplicationApiKeys
