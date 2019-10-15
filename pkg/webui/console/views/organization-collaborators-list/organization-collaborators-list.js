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
import { Container, Row, Col } from 'react-grid-system'
import bind from 'autobind-decorator'

import CollaboratorsTable from '../../containers/collaborators-table'

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

import PAGE_SIZES from '../../constants/page-sizes'

class OrganizationCollaboratorsList extends React.Component {
  static propTypes = {
    getOrganizationCollaboratorsList: PropTypes.func.isRequired,
    orgId: PropTypes.string.isRequired,
  }

  @bind
  baseDataSelector({ collaborators }) {
    const { orgId } = this.props
    return collaborators.organizations[orgId] || {}
  }

  render() {
    const { getOrganizationCollaboratorsList } = this.props

    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.collaborators} />
          <Col>
            <CollaboratorsTable
              pageSize={PAGE_SIZES.REGULAR}
              baseDataSelector={this.baseDataSelector}
              getItemsAction={getOrganizationCollaboratorsList}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}

export default OrganizationCollaboratorsList
