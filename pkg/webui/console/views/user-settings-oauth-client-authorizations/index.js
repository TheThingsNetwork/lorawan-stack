// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import AuthorizationOverview from '@console/views/user-settings-oauth-auths'
import AuthorizationsList from '@console/views/user-settings-oauth-auth-list'

import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const OAuthClientAuthorizations = () => {
  useBreadcrumbs(
    'user-settings.oauth-client-authorizations',
    <Breadcrumb
      path="/user-settings/authorizations"
      content={sharedMessages.oauthClientAuthorizations}
    />,
  )

  return (
    <Routes>
      <Route index Component={AuthorizationsList} />
      <Route
        path=":clientId/*"
        element={
          <ValidateRouteParam
            check={{ clientId: pathIdRegexp }}
            Component={AuthorizationOverview}
          />
        }
      />
      <Route path="*" element={<GenericNotFound />} />
    </Routes>
  )
}

export default OAuthClientAuthorizations
