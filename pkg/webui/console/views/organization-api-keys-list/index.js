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

import PAGE_SIZES from '@console/constants/page-sizes'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import ApiKeysTable from '@console/containers/api-keys-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getApiKeysList } from '@console/store/actions/api-keys'

import {
  selectApiKeys,
  selectApiKeysTotalCount,
  selectApiKeysFetching,
  selectApiKeysError,
} from '@console/store/selectors/api-keys'

class OrganizationApiKeysList extends React.Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }

  constructor(props) {
    super(props)

    const { orgId } = props.match.params
    this.getApiKeysList = filters => getApiKeysList('organization', orgId, filters)
  }

  @bind
  baseDataSelector(state) {
    const { orgId } = this.props.match.params

    const id = { id: orgId }
    return {
      keys: selectApiKeys(state, id),
      totalCount: selectApiKeysTotalCount(state, id),
      fetching: selectApiKeysFetching(state),
      error: selectApiKeysError(state),
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

export default OrganizationApiKeysList
