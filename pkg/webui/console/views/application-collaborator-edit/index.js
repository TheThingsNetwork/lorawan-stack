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
import { Container, Col, Row } from 'react-grid-system'
import { useParams } from 'react-router-dom'

import { APPLICATION } from '@console/constants/entities'

import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import ConsoleCollaboratorsForm from '@console/containers/collaborators-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getCollaborator, getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

const ApplicationCollaboratorEditInner = () => {
  const { appId, collaboratorId } = useParams()

  return (
    <Container>
      <PageTitle title={sharedMessages.collaboratorEdit} values={{ collaboratorId }} />
      <Row>
        <Col lg={8} md={12}>
          <ConsoleCollaboratorsForm
            entity={APPLICATION}
            entityId={appId}
            collaboratorId={collaboratorId}
            collaboratorType="user"
            update
          />
        </Col>
      </Row>
    </Container>
  )
}

const ApplicationCollaboratorEdit = () => {
  const { appId, collaboratorId, collaboratorType } = useParams()

  if (collaboratorType !== 'user' && collaboratorType !== 'organization') {
    return <GenericNotFound />
  }

  return (
    <RequireRequest
      requestAction={[
        getCollaborator('application', appId, collaboratorId, collaboratorType === 'user'),
        getCollaboratorsList('application', appId),
      ]}
    >
      <ApplicationCollaboratorEditInner />
    </RequireRequest>
  )
}

export default ApplicationCollaboratorEdit
