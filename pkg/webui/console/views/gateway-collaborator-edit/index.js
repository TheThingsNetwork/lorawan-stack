// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import { GATEWAY } from '@console/constants/entities'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import ConsoleCollaboratorsForm from '@console/containers/collaborators-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getCollaborator, getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import { selectCollaboratorById } from '@ttn-lw/lib/store/selectors/collaborators'

const GatewayCollaboratorEditInner = () => {
  const { gtwId, collaboratorId, collaboratorType } = useParams()

  useBreadcrumbs(
    'gtws.single.collaborators.edit',
    <Breadcrumb
      path={`/gateways/${gtwId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.collaboratorEdit} values={{ collaboratorId }} />
      <div className="item-12">
        <ConsoleCollaboratorsForm
          entity={GATEWAY}
          entityId={gtwId}
          collaboratorId={collaboratorId}
          collaboratorType={collaboratorType}
          update
        />
      </div>
    </div>
  )
}

const GatewayCollaboratorEdit = () => {
  const { gtwId, collaboratorId, collaboratorType } = useParams()

  const isUser = collaboratorType === 'user'

  // Check if collaborator still exists after being possibly deleted.
  const collaborator = useSelector(state => selectCollaboratorById(state, collaboratorId))
  const hasCollaborator = Boolean(collaborator)

  return (
    <RequireRequest
      requestAction={[
        getCollaborator('gateway', gtwId, collaboratorId, isUser),
        getCollaboratorsList('gateway', gtwId),
      ]}
    >
      {hasCollaborator && <GatewayCollaboratorEditInner collaboratorType={collaboratorType} />}
    </RequireRequest>
  )
}

export default GatewayCollaboratorEdit
