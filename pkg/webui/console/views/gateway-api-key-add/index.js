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

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import { ApiKeyCreateForm } from '../../components/api-key-form'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayRightsError,
  selectGatewayRightsFetching,
  selectGatewayPseudoRights,
} from '../../store/selectors/gateways'
import { mayViewOrEditGatewayApiKeys } from '../../lib/feature-checks'

import api from '../../api'
import PropTypes from '../../../lib/prop-types'

@connect(
  (state, props) => ({
    gtwId: selectSelectedGatewayId(state),
    fetching: selectGatewayRightsFetching(state),
    error: selectGatewayRightsError(state),
    rights: selectGatewayRights(state),
    pseudoRights: selectGatewayPseudoRights(state),
  }),
  dispatch => ({
    navigateToList: gtwId => dispatch(replace(`/gateways/${gtwId}/api-keys`)),
  }),
)
@withFeatureRequirement(mayViewOrEditGatewayApiKeys, {
  redirect: ({ gtwId }) => `/gateway/${gtwId}`,
})
@withBreadcrumb('gtws.single.api-keys.add', function(props) {
  const gtwId = props.gtwId

  return (
    <Breadcrumb path={`/gateways/${gtwId}/api-keys/add`} icon="add" content={sharedMessages.add} />
  )
})
@bind
export default class GatewayApiKeyAdd extends React.Component {
  constructor(props) {
    super(props)

    this.createGatewayKey = key => api.gateway.apiKeys.create(props.gtwId, key)
  }

  static propTypes = {
    gtwId: PropTypes.string.isRequired,
    navigateToList: PropTypes.func.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  handleApprove() {
    const { navigateToList, gtwId } = this.props

    navigateToList(gtwId)
  }

  render() {
    const { rights, pseudoRights } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.addApiKey} />
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyCreateForm
              rights={rights}
              pseudoRights={pseudoRights}
              onCreate={this.createGatewayKey}
              onCreateSuccess={this.handleApprove}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
