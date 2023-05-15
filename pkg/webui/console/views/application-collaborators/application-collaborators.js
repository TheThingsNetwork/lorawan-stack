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

import ErrorView from '@ttn-lw/lib/components/error-view'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import ApplicationCollaboratorsList from '@console/views/application-collaborators-list'
import ApplicationCollaboratorEdit from '@console/views/application-collaborator-edit'
import SubViewError from '@console/views/sub-view-error'
import ApplicationCollaboratorAdd from '@console/views/application-collaborator-add'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userPathId as userPathIdRegexp } from '@ttn-lw/lib/regexp'

const ApplicationCollaborators = props => {
  const { appId, match } = props

  useBreadcrumbs(
    'apps.single.collaborators',
    <Breadcrumb
      path={`/applications/${appId}/collaborators`}
      content={sharedMessages.collaborators}
    />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route exact path={`${match.path}`} component={ApplicationCollaboratorsList} />
        <Route exact path={`${match.path}/add`} component={ApplicationCollaboratorAdd} />
        <Route
          path={`${match.path}/:collaboratorType(user|organization)/:collaboratorId${userPathIdRegexp}`}
          component={ApplicationCollaboratorEdit}
          sensitive
        />
        <NotFoundRoute />
      </Routes>
    </ErrorView>
  )
}

ApplicationCollaborators.propTypes = {
  appId: PropTypes.string.isRequired,
  match: PropTypes.match.isRequired,
}

export default ApplicationCollaborators
