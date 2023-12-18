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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import GatewayApiKeyEdit from '@console/views/gateway-api-key-edit'
import GatewayApiKeyAdd from '@console/views/gateway-api-key-add'
import GatewayApiKeysList from '@console/views/gateway-api-keys-list'
import SubViewError from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { apiKeyPath as apiKeyPathRegexp } from '@console/lib/regexp'
import { mayViewOrEditGatewayApiKeys } from '@console/lib/feature-checks'

const GatewayApiKeys = () => {
  const { gtwId } = useParams()

  useBreadcrumbs('gtws.single.api-keys', [
    {
      path: `/gateways/${gtwId}/api-keys`,
      content: sharedMessages.apiKeys,
    },
  ])

  return (
    <Require
      featureCheck={mayViewOrEditGatewayApiKeys}
      otherwise={{ redirect: `/gateways/${gtwId}` }}
    >
      <ErrorView errorRender={SubViewError}>
        <Routes>
          <Route index Component={GatewayApiKeysList} />
          <Route path="add" Component={GatewayApiKeyAdd} />
          <Route
            path=":apiKeyId/*"
            element={
              <ValidateRouteParam
                check={{ apiKeyId: apiKeyPathRegexp }}
                Component={GatewayApiKeyEdit}
              />
            }
          />
          <Route path="*" Component={GenericNotFound} />
        </Routes>
      </ErrorView>
    </Require>
  )
}

export default GatewayApiKeys
