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

import IntlHelmet from '../../../lib/components/intl-helmet'
import CollaboratorsTable from '../../containers/collaborators-table'
import { getApplicationCollaboratorsList } from '../../store/actions/applications'
import sharedMessages from '../../../lib/shared-messages'

const COLLABORATORS_TABLE_SIZE = 10

@bind
export default class ApplicationCollaborators extends React.Component {

  constructor (props) {
    super(props)

    const { appId } = props.match.params
    this.getApplicationCollaboratorsList = filters => getApplicationCollaboratorsList(appId, filters)
  }

  baseDataSelector ({ collaborators }) {
    const { appId } = this.props.match.params
    return collaborators.applications[appId] || {}
  }

  render () {
    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.collaborators} />
          <Col sm={12}>
            <CollaboratorsTable
              pageSize={COLLABORATORS_TABLE_SIZE}
              baseDataSelector={this.baseDataSelector}
              getItemsAction={this.getApplicationCollaboratorsList}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}


