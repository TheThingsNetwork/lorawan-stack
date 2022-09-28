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
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import toast from '@ttn-lw/components/toast'

import CollaboratorForm from '@ttn-lw/containers/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const showSuccessToast = () => {
  toast({
    message: sharedMessages.collaboratorUpdateSuccess,
    type: toast.types.SUCCESS,
  })
}

const OrganizationCollaboratorEdit = props => {
  const {
    orgId,
    collaboratorType,
    collaboratorId,
    collaborator,
    rights,
    redirectToList,
    pseudoRights,
    updateOrganizationCollaborator,
    removeOrganizationCollaborator,
  } = props

  useBreadcrumbs(
    'orgs.single.collaborators.edit',
    <Breadcrumb
      path={`/organizations/${orgId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <Container>
      <PageTitle
        title={sharedMessages.collaboratorEdit}
        values={{ collaboratorId: collaborator.id }}
      />
      <Row>
        <Col lg={8} md={12}>
          <CollaboratorForm
            onSubmit={updateOrganizationCollaborator}
            onSubmitSuccess={showSuccessToast}
            onDelete={removeOrganizationCollaborator}
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
OrganizationCollaboratorEdit.propTypes = {
  collaborator: PropTypes.collaborator.isRequired,
  collaboratorId: PropTypes.string.isRequired,
  collaboratorType: PropTypes.string.isRequired,
  orgId: PropTypes.string.isRequired,
  pseudoRights: PropTypes.rights,
  redirectToList: PropTypes.func.isRequired,
  removeOrganizationCollaborator: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
  updateOrganizationCollaborator: PropTypes.func.isRequired,
}

OrganizationCollaboratorEdit.defaultProps = {
  pseudoRights: [],
}

export default OrganizationCollaboratorEdit
