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

import OAuthClientAdd from '@console/views/user-settings-oauth-client-add'
import OAuthClient from '@console/views/user-settings-oauth-client'
import ClientsList from '@console/views/user-settings-oauth-clients-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

const OAuthClients = () => {
  useBreadcrumbs(
    'user-settings.oauth-clients',
    <Breadcrumb path="/user-settings/oauth-clients" content={sharedMessages.oauthClients} />,
  )

  return (
    <Routes>
      <Route index Component={ClientsList} />
      <Route path="add" Component={OAuthClientAdd} />
      <Route
        path=":clientId/*"
        element={<ValidateRouteParam check={{ clientId: pathIdRegexp }} Component={OAuthClient} />}
      />
      <Route path="*" element={<GenericNotFound />} />
    </Routes>
  )
}
export default OAuthClients
