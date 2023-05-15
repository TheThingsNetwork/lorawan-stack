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

import GatewayCollaboratorEdit from '@console/views/gateway-collaborator-edit'
import GatewayCollaboratorAdd from '@console/views/gateway-collaborator-add'
import GatewayCollaboratorsList from '@console/views/gateway-collaborators-list'
import SubViewError from '@console/views/sub-view-error'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userPathId as userPathIdRegexp } from '@ttn-lw/lib/regexp'

const GatewayCollaborators = props => {
  const { match, gtwId } = props

  useBreadcrumbs(
    'gtws.single.collaborators',
    <Breadcrumb path={`/gateways/${gtwId}/collaborators`} content={sharedMessages.collaborators} />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route exact path={`${match.path}`} component={GatewayCollaboratorsList} />
        <Route exact path={`${match.path}/add`} component={GatewayCollaboratorAdd} />
        <Route
          path={`${match.path}/:collaboratorType(user|organization)/:collaboratorId${userPathIdRegexp}`}
          component={GatewayCollaboratorEdit}
        />
        <NotFoundRoute />
      </Routes>
    </ErrorView>
  )
}

GatewayCollaborators.propTypes = {
  gtwId: PropTypes.string.isRequired,
  match: PropTypes.match.isRequired,
}

export default GatewayCollaborators
