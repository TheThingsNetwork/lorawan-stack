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
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { Container, Col, Row } from 'react-grid-system'
import { replace } from 'connected-react-router'

import api from '@console/api'

import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import withRequest from '@ttn-lw/lib/components/with-request'

import { ApiKeyEditForm } from '@console/components/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getApiKey } from '@console/store/actions/api-keys'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsError,
  selectGatewayRightsFetching,
} from '@console/store/selectors/gateways'
import {
  selectSelectedApiKey,
  selectApiKeyError,
  selectApiKeyFetching,
} from '@console/store/selectors/api-keys'

@connect(
  function(state, props) {
    const apiKeyId = props.match.params.apiKeyId
    const keyFetching = selectApiKeyFetching(state)
    const rightsFetching = selectGatewayRightsFetching(state)
    const keyError = selectApiKeyError(state)
    const rightsError = selectGatewayRightsError(state)

    return {
      keyId: apiKeyId,
      gtwId: selectSelectedGatewayId(state),
      apiKey: selectSelectedApiKey(state),
      rights: selectGatewayRights(state),
      pseudoRights: selectGatewayPseudoRights(state),
      fetching: keyFetching || rightsFetching,
      error: keyError || rightsError,
    }
  },
  dispatch => ({
    getApiKey(gtwId, apiKeyId) {
      dispatch(getApiKey('gateway', gtwId, apiKeyId))
    },
    deleteSuccess: gtwId => dispatch(replace(`/gateways/${gtwId}/api-keys`)),
  }),
)
@withRequest(
  ({ gtwId, keyId, getApiKey }) => getApiKey(gtwId, keyId),
  ({ fetching, apiKey }) => fetching || !Boolean(apiKey),
)
@withBreadcrumb('gtws.single.api-keys.edit', function(props) {
  const { gtwId, keyId } = props

  return <Breadcrumb path={`/gateways/${gtwId}/api-keys/${keyId}`} content={sharedMessages.edit} />
})
@bind
export default class GatewayApiKeyEdit extends React.Component {
  static propTypes = {
    apiKey: PropTypes.apiKey.isRequired,
    deleteSuccess: PropTypes.func.isRequired,
    gtwId: PropTypes.string.isRequired,
    keyId: PropTypes.string.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  constructor(props) {
    super(props)

    this.deleteGatewayKey = id => api.gateway.apiKeys.delete(props.gtwId, id)
    this.editGatewayKey = key => api.gateway.apiKeys.update(props.gtwId, props.keyId, key)
  }

  onDeleteSuccess() {
    const { gtwId, deleteSuccess } = this.props

    deleteSuccess(gtwId)
  }

  render() {
    const { apiKey, rights, pseudoRights } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.keyEdit} />
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyEditForm
              rights={rights}
              pseudoRights={pseudoRights}
              apiKey={apiKey}
              onEdit={this.editGatewayKey}
              onDelete={this.deleteGatewayKey}
              onDeleteSuccess={this.onDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
