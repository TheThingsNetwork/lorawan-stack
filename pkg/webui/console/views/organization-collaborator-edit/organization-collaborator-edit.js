// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import PageTitle from '../../../components/page-title'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import CollaboratorForm from '../../components/collaborator-form'
import toast from '../../../components/toast'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

@withBreadcrumb('orgs.single.collaborators.edit', function(props) {
  const { orgId, collaboratorId, collaboratorType } = props

  return (
    <Breadcrumb
      path={`/organizations/${orgId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />
  )
})
class OrganizationCollaboratorEdit extends React.Component {
  static propTypes = {
    collaborator: PropTypes.collaborator.isRequired,
    pseudoRights: PropTypes.rights,
    redirectToList: PropTypes.func.isRequired,
    removeOrganizationCollaborator: PropTypes.func.isRequired,
    rights: PropTypes.rights.isRequired,
    updateOrganizationCollaborator: PropTypes.func.isRequired,
  }

  static defaultProps = {
    pseudoRights: [],
  }

  handleSubmitSuccess() {
    toast({
      message: sharedMessages.collaboratorUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  render() {
    const {
      collaborator,
      rights,
      redirectToList,
      pseudoRights,
      updateOrganizationCollaborator,
      removeOrganizationCollaborator,
    } = this.props

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
              onSubmitSuccess={this.handleSubmitSuccess}
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
}

export default OrganizationCollaboratorEdit
