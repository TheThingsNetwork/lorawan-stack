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

import React from 'react'
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import { ApiKeyEditForm } from '@console/components/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const GatewayApiKeyEdit = props => {
  const {
    editGatewayApiKey,
    apiKey,
    deleteGatewayApiKey,
    gtwId,
    pseudoRights,
    rights,
    keyId,
    deleteSuccess,
  } = props

  useBreadcrumbs(
    'gtws.single.api-keys.edit',
    <Breadcrumb path={`/gateways/${gtwId}/api-keys/${keyId}`} content={sharedMessages.edit} />,
  )

  return (
    <Container>
      <PageTitle title={sharedMessages.keyEdit} />
      <Row>
        <Col lg={8} md={12}>
          <ApiKeyEditForm
            rights={rights}
            pseudoRights={pseudoRights}
            apiKey={apiKey}
            onEdit={editGatewayApiKey}
            onDelete={deleteGatewayApiKey}
            onDeleteSuccess={deleteSuccess}
          />
        </Col>
      </Row>
    </Container>
  )
}

GatewayApiKeyEdit.propTypes = {
  apiKey: PropTypes.apiKey.isRequired,
  deleteGatewayApiKey: PropTypes.func.isRequired,
  deleteSuccess: PropTypes.func.isRequired,
  editGatewayApiKey: PropTypes.func.isRequired,
  gtwId: PropTypes.string.isRequired,
  keyId: PropTypes.string.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  rights: PropTypes.rights.isRequired,
}

export default GatewayApiKeyEdit
