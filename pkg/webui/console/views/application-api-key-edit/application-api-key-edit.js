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

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import { ApiKeyEditForm } from '@console/components/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

@withBreadcrumb('apps.single.api-keys.edit', props => {
  const { appId, keyId } = props

  return (
    <Breadcrumb path={`/applications/${appId}/api-keys/${keyId}`} content={sharedMessages.edit} />
  )
})
export default class ApplicationApiKeyEdit extends React.Component {
  static propTypes = {
    apiKey: PropTypes.apiKey.isRequired,
    deleteApplicationKey: PropTypes.func.isRequired,
    deleteSuccess: PropTypes.func.isRequired,
    editApplicationKey: PropTypes.func.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  @bind
  onDeleteSuccess() {
    const { deleteSuccess } = this.props

    deleteSuccess()
  }

  render() {
    const { apiKey, rights, pseudoRights, deleteApplicationKey, editApplicationKey } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.keyEdit} />
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyEditForm
              rights={rights}
              pseudoRights={pseudoRights}
              apiKey={apiKey}
              onEdit={editApplicationKey}
              onDelete={deleteApplicationKey}
              onDeleteSuccess={this.onDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
