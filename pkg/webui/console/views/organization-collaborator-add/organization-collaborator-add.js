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
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import CollaboratorForm from '@ttn-lw/containers/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const OrganizationCollaboratorAdd = props => {
  const { orgId, rights, redirectToList, pseudoRights, addOrganizationCollaborator } = props

  useBreadcrumbs(
    'orgs.single.collaborators.add',
    <Breadcrumb path={`/organizations/${orgId}/collaborators/add`} content={sharedMessages.add} />,
  )

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

OrganizationCollaboratorAdd.propTypes = {
  addOrganizationCollaborator: PropTypes.func.isRequired,
  orgId: PropTypes.string.isRequired,
  pseudoRights: PropTypes.rights,
  redirectToList: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
}

OrganizationCollaboratorAdd.defaultProps = {
  pseudoRights: [],
}

export default OrganizationCollaboratorAdd
