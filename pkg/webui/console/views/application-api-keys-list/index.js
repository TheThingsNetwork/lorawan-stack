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
import { getApplicationApiKeysList } from '../../store/actions/application'
import sharedMessages from '../../../lib/shared-messages'

const API_KEYS_TABLE_SIZE = 10

@bind
export default class ApplicationApiKeys extends React.Component {

  constructor (props) {
    super(props)

    const { appId } = props.match.params
    this.getApplicationsApiKeysList = filters => getApplicationApiKeysList(appId, filters)
  }

  baseDataSelector ({ apiKeys }) {
    const { appId } = this.props.match.params
    return apiKeys.applications[appId] || {}
  }

  render () {
    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.apiKeys} />
          <Col sm={12}>
            <ApiKeysTable
              pageSize={API_KEYS_TABLE_SIZE}
              baseDataSelector={this.baseDataSelector}
              getItemsAction={this.getApplicationsApiKeysList}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
