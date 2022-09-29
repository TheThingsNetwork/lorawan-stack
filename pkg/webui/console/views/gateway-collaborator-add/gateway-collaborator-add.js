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
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import CollaboratorForm from '@ttn-lw/containers/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const GatewayCollaboratorAdd = props => {
  const { gtwId, rights, redirectToList, pseudoRights, addCollaborator, error } = props

  useBreadcrumbs(
    'gtws.single.collaborators.add',
    <Breadcrumb path={`/gateways/${gtwId}/collaborators/add`} content={sharedMessages.add} />,
  )

  const handleSubmit = useCallback(collaborator => addCollaborator(collaborator), [addCollaborator])

  return (
    <Container>
      <PageTitle title={sharedMessages.addCollaborator} />
      <Row>
        <Col lg={8} md={12}>
          <CollaboratorForm
            error={error}
            onSubmit={handleSubmit}
            onSubmitSuccess={redirectToList}
            pseudoRights={pseudoRights}
            rights={rights}
          />
        </Col>
      </Row>
    </Container>
  )
}

GatewayCollaboratorAdd.propTypes = {
  addCollaborator: PropTypes.func.isRequired,
  error: PropTypes.error,
  gtwId: PropTypes.string.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  redirectToList: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
}

GatewayCollaboratorAdd.defaultProps = {
  error: undefined,
}
export default GatewayCollaboratorAdd
