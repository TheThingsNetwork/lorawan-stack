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

import ErrorView from '@ttn-lw/lib/components/error-view'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import SubViewError from '@console/views/sub-view-error'
import OrganizationApiKeysList from '@console/views/organization-api-keys-list'
import OrganizationApiKeyAdd from '@console/views/organization-api-key-add'
import OrganizationApiKeyEdit from '@console/views/organization-api-key-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { apiKeyPath as apiKeyPathRegexp } from '@console/lib/regexp'

const OrganizationApiKeys = props => {
  const { orgId, match } = props

  useBreadcrumbs(
    'org.single.api-keys',
    <Breadcrumb path={`/organizations/${orgId}/api-keys`} content={sharedMessages.apiKeys} />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route exact path={`${match.path}`} component={OrganizationApiKeysList} />
        <Route exact path={`${match.path}/add`} component={OrganizationApiKeyAdd} />
        <Route
          path={`${match.path}/:apiKeyId${apiKeyPathRegexp}`}
          component={OrganizationApiKeyEdit}
          sensitive
        />
        <NotFoundRoute />
      </Routes>
    </ErrorView>
  )
}

OrganizationApiKeys.propTypes = {
  match: PropTypes.match.isRequired,
  orgId: PropTypes.string.isRequired,
}

export default OrganizationApiKeys
