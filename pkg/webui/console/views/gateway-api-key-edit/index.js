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

import PageTitle from '../../../components/page-title'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import { ApiKeyEditForm } from '../../components/api-key-form'
import withRequest from '../../../lib/components/with-request'

import { getGatewayApiKey } from '../../store/actions/gateways'
import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsError,
  selectGatewayRightsFetching,
  selectGatewayApiKey,
  selectGatewayApiKeyError,
  selectGatewayApiKeyFetching,
} from '../../store/selectors/gateways'

import api from '../../api'
import PropTypes from '../../../lib/prop-types'

@connect(
  function(state, props) {
    const apiKeyId = props.match.params.apiKeyId
    const keyFetching = selectGatewayApiKeyFetching(state)
    const rightsFetching = selectGatewayRightsFetching(state)
    const keyError = selectGatewayApiKeyError(state)
    const rightsError = selectGatewayRightsError(state)

    return {
      keyId: apiKeyId,
      gtwId: selectSelectedGatewayId(state),
      apiKey: selectGatewayApiKey(state),
      rights: selectGatewayRights(state),
      pseudoRights: selectGatewayPseudoRights(state),
      fetching: keyFetching || rightsFetching,
      error: keyError || rightsError,
    }
  },
  dispatch => ({
    getGatewayApiKey(gtwId, apiKeyId) {
      dispatch(getGatewayApiKey(gtwId, apiKeyId))
    },
    deleteSuccess: gtwId => dispatch(replace(`/gateways/${gtwId}/api-keys`)),
  }),
)
@withRequest(
  ({ gtwId, keyId, getGatewayApiKey }) => getGatewayApiKey(gtwId, keyId),
  ({ fetching, apiKey }) => fetching || !Boolean(apiKey),
)
@withBreadcrumb('gtws.single.api-keys.edit', function(props) {
  const { gtwId, keyId } = props

  return (
    <Breadcrumb
      path={`/gateways/${gtwId}/api-keys/${keyId}`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
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
