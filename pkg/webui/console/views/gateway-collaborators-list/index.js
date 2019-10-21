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
import { connect } from 'react-redux'

import IntlHelmet from '../../../lib/components/intl-helmet'
import CollaboratorsTable from '../../containers/collaborators-table'
import sharedMessages from '../../../lib/shared-messages'

import { getGatewayCollaboratorsList } from '../../store/actions/gateways'
import {
  selectSelectedGatewayId,
  selectGatewayCollaborators,
  selectGatewayCollaboratorsTotalCount,
  selectGatewayCollaboratorsFetching,
} from '../../store/selectors/gateways'

import PAGE_SIZES from '../../constants/page-sizes'

@connect(state => ({
  gtwId: selectSelectedGatewayId(state),
}))
@bind
export default class GatewayCollaborators extends React.Component {
  constructor(props) {
    super(props)

    const { gtwId } = this.props
    this.getGatewayCollaboratorsList = filters => getGatewayCollaboratorsList(gtwId, filters)
  }

  baseDataSelector(state) {
    const { gtwId } = this.props
    const id = { id: gtwId }

    return {
      collaborators: selectGatewayCollaborators(state, id),
      fetching: selectGatewayCollaboratorsFetching(state),
      totalCount: selectGatewayCollaboratorsTotalCount(state, id),
    }
  }

  render() {
    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.collaborators} />
          <Col>
            <CollaboratorsTable
              pageSize={PAGE_SIZES.REGULAR}
              baseDataSelector={this.baseDataSelector}
              getItemsAction={this.getGatewayCollaboratorsList}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
