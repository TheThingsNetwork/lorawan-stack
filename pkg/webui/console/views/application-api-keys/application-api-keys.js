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

import ApplicationApiKeysList from '@console/views/application-api-keys-list'
import ApplicationApiKeyAdd from '@console/views/application-api-key-add'
import SubViewError from '@console/views/sub-view-error'
import ApplicationApiKeyEdit from '@console/views/application-api-key-edit'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { apiKeyPath as apiKeyPathRegexp } from '@console/lib/regexp'

const ApplicationApiKeys = props => {
  const { appId, match } = props

  useBreadcrumbs(
    'apps.single.api-keys',
    <Breadcrumb path={`/applications/${appId}/api-keys`} content={sharedMessages.apiKeys} />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route exact path={`${match.path}`} component={ApplicationApiKeysList} />
        <Route exact path={`${match.path}/add`} component={ApplicationApiKeyAdd} />
        <Route
          path={`${match.path}/:apiKeyId${apiKeyPathRegexp}`}
          component={ApplicationApiKeyEdit}
          sensitive
        />
        <NotFoundRoute />
      </Routes>
    </ErrorView>
  )
}

ApplicationApiKeys.propTypes = {
  appId: PropTypes.string.isRequired,
  match: PropTypes.match.isRequired,
}

export default ApplicationApiKeys
