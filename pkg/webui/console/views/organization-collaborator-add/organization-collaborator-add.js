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
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import CollaboratorForm from '../../components/collaborator-form'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

@withBreadcrumb('orgs.single.collaborators.add', props => {
  const { orgId } = props

  return (
    <Breadcrumb
      path={`/organizations/${orgId}/collaborators/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
class OrganizationCollaboratorAdd extends React.Component {
  static propTypes = {
    addOrganizationCollaborator: PropTypes.func.isRequired,
    pseudoRights: PropTypes.rights,
    redirectToList: PropTypes.func.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  static defaultProps = {
    pseudoRights: [],
  }

  render() {
    const { rights, redirectToList, pseudoRights, addOrganizationCollaborator } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.addCollaborator} />
        <Row>
          <Col lg={8} md={12}>
            <CollaboratorForm
              onSubmit={addOrganizationCollaborator}
              onSubmitSuccess={redirectToList}
              pseudoRights={pseudoRights}
              rights={rights}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}

export default OrganizationCollaboratorAdd
