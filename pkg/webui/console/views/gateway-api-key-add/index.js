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
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { replace } from 'connected-react-router'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { ApiKeyCreateForm } from '../../components/api-key-form'
import withRequest from '../../../lib/components/with-request'

import { getGatewaysRightsList } from '../../store/actions/gateways'
import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayRightsError,
  selectGatewayRightsFetching,
  selectGatewayUniversalRights,
} from '../../store/selectors/gateways'

import api from '../../api'

@connect((state, props) => ({
  gtwId: selectSelectedGatewayId(state),
  fetching: selectGatewayRightsFetching(state),
  error: selectGatewayRightsError(state),
  rights: selectGatewayRights(state),
  universalRights: selectGatewayUniversalRights(state),
}),
dispatch => ({
  getGatewaysRightsList: gtwId => dispatch(getGatewaysRightsList(gtwId)),
  navigateToList: gtwId => dispatch(replace(`/gateways/${gtwId}/api-keys`)),
}))
@withRequest(
  ({ gtwId, getGatewaysRightsList }) => getGatewaysRightsList(gtwId),
  ({ fetching, rights }) => fetching || !Boolean(rights.length)
)
@withBreadcrumb('gtws.single.api-keys.add', function (props) {
  const gtwId = props.gtwId

  return (
    <Breadcrumb
      path={`/gateways/${gtwId}/api-keys/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class GatewayApiKeyAdd extends React.Component {

  constructor (props) {
    super(props)

    this.createGatewayKey = key => api.gateway.apiKeys.create(props.gtwId, key)
  }

  handleApprove () {
    const { navigateToList, gtwId } = this.props

    navigateToList(gtwId)
  }

  render () {
    const { rights, universalRights } = this.props

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.addApiKey} />
            <Message component="h2" content={sharedMessages.addApiKey} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyCreateForm
              rights={rights}
              universalRights={universalRights}
              onCreate={this.createGatewayKey}
              onCreateSuccess={this.handleApprove}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
