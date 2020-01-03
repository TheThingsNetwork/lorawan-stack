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
import { getApiKeysList } from '../../store/actions/api-keys'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

import {
  selectApiKeys,
  selectApiKeysTotalCount,
  selectApiKeysFetching,
} from '../../store/selectors/api-keys'

import PAGE_SIZES from '../../constants/page-sizes'

export default class ApplicationApiKeys extends React.Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }

  constructor(props) {
    super(props)

    const { appId } = props.match.params
    this.getApiKeysList = filters => getApiKeysList('application', appId, filters)
  }

  @bind
  baseDataSelector(state) {
    const { appId } = this.props.match.params

    const id = { id: appId }
    return {
      keys: selectApiKeys(state, id),
      totalCount: selectApiKeysTotalCount(state, id),
      fetching: selectApiKeysFetching(state),
    }
  }

  render() {
    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.apiKeys} />
          <Col>
            <ApiKeysTable
              pageSize={PAGE_SIZES.REGULAR}
              baseDataSelector={this.baseDataSelector}
              getItemsAction={this.getApiKeysList}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
