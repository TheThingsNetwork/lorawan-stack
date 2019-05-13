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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { ApiKeyEditForm } from '../../../components/api-key-form'

import { getGatewayApiKeyPageData } from '../../store/actions/gateway'
import { getGatewayId } from '../../../lib/selectors/id'
import {
  gatewaySelector,
  gatewayRightsSelector,
  gatewayRightsErrorSelector,
  gatewayRightsFetchingSelector,
  gatewayKeySelector,
  gatewayKeysErrorSelector,
  gatewayKeysFetchingSelector,
} from '../../store/selectors/gateway'

import api from '../../api'

@connect(function (state, props) {
  const gateway = gatewaySelector(state, props)
  const gtwId = getGatewayId(gateway)
  const apiKeyId = props.match.params.apiKeyId

  const ids = { id: gtwId, keyId: apiKeyId }

  const keysFetching = gatewayKeysFetchingSelector(state, ids)
  const rightsFetching = gatewayRightsFetchingSelector(state, props)
  const keysError = gatewayKeysErrorSelector(state, ids)
  const apiKey = gatewayKeySelector(state, ids)
  const rightsError = gatewayRightsErrorSelector(state, props)
  const rights = gatewayRightsSelector(state, props)

  return {
    keyId: apiKeyId,
    gtwId,
    apiKey,
    rights,
    fetching: keysFetching || rightsFetching,
    error: keysError || rightsError,
  }
})
@withBreadcrumb('gtws.single.api-keys.edit', function (props) {
  const { gtwId, keyId } = props

  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}/api-keys/${keyId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class GatewayApiKeyEdit extends React.Component {

  constructor (props) {
    super(props)

    this.deleteGatewayKey = id => api.gateway.apiKeys.delete(props.gtwId, id)
    this.editGatewayKey = key => api.gateway.apiKeys.update(
      props.gtwId,
      props.apiKey.id,
      key
    )
  }

  componentDidMount () {
    const { dispatch, gtwId } = this.props

    dispatch(getGatewayApiKeyPageData(gtwId))
  }

  onDeleteSuccess () {
    const { gtwId, dispatch } = this.props

    dispatch(replace(`/console/gateways/${gtwId}/api-keys`))
  }

  render () {
    const { apiKey, rights, fetching, error } = this.props

    if (error) {
      return 'ERROR'
    }

    if (fetching || !apiKey) {
      return <Spinner center />
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.keyEdit} />
            <Message component="h2" content={sharedMessages.keyEdit} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyEditForm
              rights={rights}
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
