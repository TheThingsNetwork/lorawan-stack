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

import React, { useCallback } from 'react'
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import toast from '@ttn-lw/components/toast'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import CollaboratorForm from '@ttn-lw/containers/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const handleSubmitSuccess = () => {
  toast({
    message: sharedMessages.collaboratorUpdateSuccess,
    type: toast.types.SUCCESS,
  })
}

const GatewayCollaboratorEdit = props => {
  const {
    collaborator,
    collaboratorId,
    collaboratorType,
    gtwId,
    pseudoRights,
    redirectToList,
    removeCollaborator,
    rights,
    updateCollaborator,
  } = props

  useBreadcrumbs(
    'gtws.single.collaborators.edit',
    <Breadcrumb
      path={`/gateways/${gtwId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />,
  )

  const handleSubmit = useCallback(patch => updateCollaborator(patch), [updateCollaborator])
  const handleDelete = useCallback(
    collaboratorIds => removeCollaborator(collaboratorIds),
    [removeCollaborator],
  )

  return (
    <Container>
      <PageTitle title={sharedMessages.collaboratorEdit} values={{ collaboratorId }} />
      <Row>
        <Col lg={8} md={12}>
          <CollaboratorForm
            onSubmit={handleSubmit}
            onSubmitSuccess={handleSubmitSuccess}
            onDelete={handleDelete}
            onDeleteSuccess={redirectToList}
            collaborator={collaborator}
            pseudoRights={pseudoRights}
            rights={rights}
            update
          />
        </Col>
      </Row>
    </Container>
  )
}

GatewayCollaboratorEdit.propTypes = {
  collaborator: PropTypes.collaborator.isRequired,
  collaboratorId: PropTypes.string.isRequired,
  collaboratorType: PropTypes.string.isRequired,
  gtwId: PropTypes.string.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  redirectToList: PropTypes.func.isRequired,
  removeCollaborator: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
  updateCollaborator: PropTypes.func.isRequired,
}

export default GatewayCollaboratorEdit
