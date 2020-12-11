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

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import { ApiKeyCreateForm } from '@console/components/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const UserApiKeyAdd = props => {
  const { navigateToList, rights, pseudoRights, createUserApiKey } = props

  return (
    <Container>
      <PageTitle title={sharedMessages.addApiKey} />
      <Row>
        <Col lg={8} md={12}>
          <ApiKeyCreateForm
            rights={rights}
            pseudoRights={pseudoRights}
            onCreate={createUserApiKey}
            onCreateSuccess={navigateToList}
          />
        </Col>
      </Row>
    </Container>
  )
}

UserApiKeyAdd.propTypes = {
  createUserApiKey: PropTypes.func.isRequired,
  navigateToList: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights,
  rights: PropTypes.rights.isRequired,
}

UserApiKeyAdd.defaultProps = {
  pseudoRights: [],
}

export default withBreadcrumb('usr.single.api-keys.add', () => (
  <Breadcrumb path={`/users/api-keys/add`} content={sharedMessages.add} />
))(UserApiKeyAdd)
