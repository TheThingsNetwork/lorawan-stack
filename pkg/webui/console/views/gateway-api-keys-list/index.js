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
import ApiKeysTable from '../../containers/api-keys-table'
import sharedMessages from '../../../lib/shared-messages'

import { getGatewayApiKeysList } from '../../store/actions/gateways'
import { selectGatewayApiKeysStore } from '../../store/selectors/gateway'

const API_KEYS_TABLE_SIZE = 10

@bind
export default class GatewayApiKeys extends React.Component {

  constructor (props) {
    super(props)

    const { gtwId } = props.match.params
    this.getGatewayApiKeysList = filters => getGatewayApiKeysList(gtwId, filters)
  }

  render () {
    const { gtwId } = this.props.match.params

    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.apiKeys} />
          <Col sm={12}>
            <ApiKeysTable
              entityId={gtwId}
              pageSize={API_KEYS_TABLE_SIZE}
              baseDataSelector={selectGatewayApiKeysStore}
              getItemsAction={this.getGatewayApiKeysList}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
