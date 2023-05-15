// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import UserApiKeyEdit from '@console/views/user-api-key-edit'
import UserApiKeyAdd from '@console/views/user-api-key-add'
import SubViewError from '@console/views/sub-view-error'
import UserApiKeysList from '@console/views/user-api-keys-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { apiKeyPath as apiKeyPathRegexp } from '@console/lib/regexp'

const UserApiKeys = ({ match }) => (
  <ErrorView errorRender={SubViewError}>
    <Routes>
      <Route exact path={match.path} component={UserApiKeysList} />
      <Route exact path={`${match.path}/add`} component={UserApiKeyAdd} />
      <Route
        path={`${match.path}/:apiKeyId${apiKeyPathRegexp}`}
        component={UserApiKeyEdit}
        sensitive
      />
      <NotFoundRoute />
    </Routes>
  </ErrorView>
)

UserApiKeys.propTypes = {
  match: PropTypes.match.isRequired,
}

export default withBreadcrumb('usr.single.api-keys', () => (
  <Breadcrumb path={`/user/api-keys`} content={sharedMessages.personalApiKeys} />
))(UserApiKeys)
