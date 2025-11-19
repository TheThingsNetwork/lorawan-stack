// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
import { useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import { CLIENT } from '@console/constants/entities'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import AccountCollaboratorsForm from '@console/containers/collaborators-form'

import { selectCollaboratorById } from '@ttn-lw/lib/store/selectors/collaborators'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getCollaborator } from '@ttn-lw/lib/store/actions/collaborators'

const OAuthClientCollaboratorEditInner = () => {
  const { clientId, collaboratorId, collaboratorType } = useParams()

  useBreadcrumbs(
    'user-settings.oauth-clients.single.collaborators.single.edit',
    <Breadcrumb
      path={`/user-settings/oauth-clients/${clientId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <div className="container container--xxl grid">
      <PageTitle title={sharedMessages.collaboratorEdit} values={{ collaboratorId }} />
      <div className="item-12 xl:item-8">
        <AccountCollaboratorsForm
          entity={CLIENT}
          entityId={clientId}
          collaboratorId={collaboratorId}
          collaboratorType={collaboratorType === 'user' ? 'user' : 'organization'}
          update
        />
      </div>
    </div>
  )
}

const OAuthClientCollaboratorEdit = () => {
  const { clientId, collaboratorId, collaboratorType } = useParams()

  // Check if collaborator still exists after being possibly deleted.
  const collaborator = useSelector(state => selectCollaboratorById(state, collaboratorId))
  const hasCollaborator = Boolean(collaborator)
  const isUser = collaboratorType === 'user'

  useBreadcrumbs(
    'user-settings.oauth-clients.single.collaborators.single',
    <Breadcrumb
      path={`/user-settings/oauth-clients/${clientId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={collaboratorId}
    />,
  )

  if (collaboratorType !== 'user' && collaboratorType !== 'organization') {
    return null
  }

  return (
    <RequireRequest requestAction={getCollaborator('client', clientId, collaboratorId, isUser)}>
      {hasCollaborator && <OAuthClientCollaboratorEditInner />}
    </RequireRequest>
  )
}

export default OAuthClientCollaboratorEdit
