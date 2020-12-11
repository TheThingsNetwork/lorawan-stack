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
import { Container, Col, Row } from 'react-grid-system'

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'

import { ApiKeyEditForm } from '@console/components/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const UserApiKeyEdit = props => {
  const {
    apiKey,
    rights,
    pseudoRights,
    deleteUserApiKey,
    deleteUserApiKeySuccess,
    updateUserApiKey,
  } = props

  return (
    <Container>
      <PageTitle title={sharedMessages.keyEdit} />
      <Row>
        <Col lg={8} md={12}>
          <ApiKeyEditForm
            rights={rights}
            pseudoRights={pseudoRights}
            apiKey={apiKey}
            onEdit={updateUserApiKey}
            onDelete={deleteUserApiKey}
            onDeleteSuccess={deleteUserApiKeySuccess}
          />
        </Col>
      </Row>
    </Container>
  )
}

UserApiKeyEdit.propTypes = {
  apiKey: PropTypes.apiKey.isRequired,
  deleteUserApiKey: PropTypes.func.isRequired,
  deleteUserApiKeySuccess: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights,
  rights: PropTypes.rights.isRequired,
  updateUserApiKey: PropTypes.func.isRequired,
}

UserApiKeyEdit.defaultProps = {
  pseudoRights: [],
}

export default withBreadcrumb('usr.single.api-keys.edit', props => {
  const { keyId } = props

  return <Breadcrumb path={`/user/api-keys/${keyId}`} content={sharedMessages.edit} />
})(UserApiKeyEdit)
