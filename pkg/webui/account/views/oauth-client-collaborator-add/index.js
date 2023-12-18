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

import { CLIENT } from '@console/constants/entities'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import AccountCollaboratorsForm from '@account/containers/collaborators-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getClientRights } from '@account/store/actions/clients'

const OAuthClientCollaboratorAddInner = () => {
  const { clientId } = useParams()
  useBreadcrumbs('clients.single.collaborators.add', [
    {
      path: `/oauth-clients/${clientId}/collaborators/add`,
      content: sharedMessages.add,
    },
  ])

  return (
    <Container>
      <PageTitle title={sharedMessages.addCollaborator} />
      <Row>
        <Col lg={8} md={12}>
          <AccountCollaboratorsForm entity={CLIENT} entityId={clientId} />
        </Col>
      </Row>
    </Container>
  )
}

const OAuthClientCollaboratorAdd = () => {
  const { clientId } = useParams()

  return (
    <RequireRequest requestAction={getClientRights(clientId)}>
      <OAuthClientCollaboratorAddInner />
    </RequireRequest>
  )
}

export default OAuthClientCollaboratorAdd
