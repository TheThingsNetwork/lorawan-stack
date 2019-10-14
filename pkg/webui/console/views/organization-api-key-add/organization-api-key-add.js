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
import bind from 'autobind-decorator'

import { ApiKeyCreateForm } from '../../components/api-key-form'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import PropTypes from '../../../lib/prop-types'

@withBreadcrumb('orgs.single.api-keys.add', function(props) {
  const orgId = props.orgId
  return (
    <Breadcrumb
      path={`/organizations/${orgId}/api-keys/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
class OrganizationApiKeyAdd extends React.Component {
  static propTypes = {
    createOrganizationApiKey: PropTypes.func.isRequired,
    navigateToList: PropTypes.func.isRequired,
    pseudoRights: PropTypes.rights,
    rights: PropTypes.rights.isRequired,
  }

  static defaultProps = {
    pseudoRights: [],
  }

  @bind
  handleApprove() {
    const { navigateToList } = this.props

    navigateToList()
  }

  render() {
    const { rights, pseudoRights, createOrganizationApiKey } = this.props

    return (
      <Container>
        <Row>
          <Col>
            <IntlHelmet title={sharedMessages.addApiKey} />
            <Message component="h2" content={sharedMessages.addApiKey} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyCreateForm
              rights={rights}
              pseudoRights={pseudoRights}
              onCreate={createOrganizationApiKey}
              onCreateSuccess={this.handleApprove}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}

export default OrganizationApiKeyAdd
