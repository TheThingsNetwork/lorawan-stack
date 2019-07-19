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
import { push } from 'connected-react-router'

import Spinner from '../../../components/spinner'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import CollaboratorForm from '../../components/collaborator-form'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayUniversalRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
} from '../../store/selectors/gateways'

import { getGatewaysRightsList } from '../../store/actions/gateways'
import api from '../../api'

@connect(function (state, props) {
  const { collaborators } = state

  return {
    gtwId: selectSelectedGatewayId(state),
    collaborators: collaborators.gateways.collaborators,
    fetching: selectGatewayRightsFetching(state),
    error: selectGatewayRightsError(state),
    rights: selectGatewayRights(state),
    universalRights: selectGatewayUniversalRights(state),
  }
}, (dispatch, ownProps) => ({
  getGatewaysRightsList: gtwId => dispatch(getGatewaysRightsList(gtwId)),
  redirectToList: gtwId => dispatch(push(`/console/gateways/${gtwId}/collaborators`)),
}), (stateProps, dispatchProps, ownProps) => ({
  ...stateProps, ...dispatchProps, ...ownProps,
  getGatewaysRightsList: () => dispatchProps.getGatewaysRightsList(stateProps.gtwId),
  redirectToList: () => dispatchProps.redirectToList(stateProps.gtwId),
}))
@withBreadcrumb('gtws.single.collaborators.add', function (props) {
  const gtwId = props.gtwId
  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}/collaborators/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class GatewayCollaboratorAdd extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { getGatewaysRightsList } = this.props

    getGatewaysRightsList()
  }

  componentDidUpdate (prevProps) {
    const { error } = this.props

    if (Boolean(error) && prevProps.error !== error) {
      throw error
    }
  }

  handleSubmit (collaborator) {
    const { gtwId } = this.props

    return api.gateway.collaborators.add(gtwId, collaborator)
  }

  render () {
    const {
      rights,
      fetching,
      redirectToList,
      universalRights,
    } = this.props

    if (fetching && !rights.length) {
      return <Spinner center />
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.addCollaborator} />
            <Message component="h2" content={sharedMessages.addCollaborator} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <CollaboratorForm
              error={this.state.error}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={redirectToList}
              universalRights={universalRights}
              rights={rights}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
