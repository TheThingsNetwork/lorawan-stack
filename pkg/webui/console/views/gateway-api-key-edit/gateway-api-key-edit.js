// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import bind from 'autobind-decorator'
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import { ApiKeyEditForm } from '@console/components/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

@withBreadcrumb('gtws.single.api-keys.edit', props => {
  const { gtwId, keyId } = props

  return <Breadcrumb path={`/gateways/${gtwId}/api-keys/${keyId}`} content={sharedMessages.edit} />
})
export default class GatewayApiKeyEdit extends React.Component {
  static propTypes = {
    apiKey: PropTypes.apiKey.isRequired,
    deleteGatewayApiKey: PropTypes.func.isRequired,
    deleteSuccess: PropTypes.func.isRequired,
    editGatewayApiKey: PropTypes.func.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  @bind
  onDeleteSuccess() {
    const { deleteSuccess } = this.props

    deleteSuccess()
  }

  render() {
    const { apiKey, rights, pseudoRights, deleteGatewayApiKey, editGatewayApiKey } = this.props

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
              onDeleteSuccess={this.onDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
