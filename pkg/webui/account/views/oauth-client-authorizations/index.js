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
import { Routes, Route } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import AuthorizationsList from '@account/views/oauth-client-authorizations-list'
import AuthorizationOverview from '@account/views/oauth-client-authorization-overview'

import PropTypes from '@ttn-lw/lib/prop-types'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const OAuthClientAuthorizations = props => {
  const { path } = props.match

  useBreadcrumbs(
    'client-authorizations',
    <Breadcrumb path="/client-authorizations" content={sharedMessages.oauthClientAuthorizations} />,
  )

  return (
    <Routes>
      <Route exact path={`${path}`} component={AuthorizationsList} />
      <Route path={`${path}/:clientId${pathIdRegexp}`} component={AuthorizationOverview} />
      <NotFoundRoute />
    </Routes>
  )
}
OAuthClientAuthorizations.propTypes = {
  match: PropTypes.match.isRequired,
}
export default OAuthClientAuthorizations
